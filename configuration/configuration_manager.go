package configuration

import (
	"sync"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"

	"github.com/Kameleoon/client-go/v3/events"
	"github.com/Kameleoon/client-go/v3/managers/data"
	"github.com/Kameleoon/client-go/v3/network"
	"github.com/Kameleoon/client-go/v3/realtime"
	"github.com/segmentio/encoding/json"
)

// Not thread-safe
type ConfigurationManager interface {
	Start() error
	OnUpdateConfiguration(handler func())
	TryFetch(ts int64) error
}

// ClientReadiness receives the outcome of every configuration fetch: a successful fetch
// marks the SDK ready, a failed one reports the failure which prevented the SDK from
// loading its configuration.
type ClientReadiness interface {
	MarkReady()
	MarkNotReady(cause error)
}

type configurationManagerImpl struct {
	dataManager    data.DataManager
	networkManager network.NetworkManager
	eventManager   events.EventManager
	sseClient      realtime.SseClient

	pollingUpdateInterval time.Duration
	environment           string

	updateConfigurationHandler func()
	readiness                  ClientReadiness

	// mx guards the data file update.
	mx sync.Mutex

	// updateModeMx serializes the switches between the streaming and the polling
	// configuration update mode and guards the fields below, so concurrent fetch
	// outcomes cannot interleave the mode transitions.
	updateModeMx sync.Mutex
	// pollingRunning is true while a polling loop is running.
	pollingRunning       bool
	realTimeEventService *realtime.RealTimeEventService
	realTimeUpdateChan   chan realtime.RealTimeEvent
}

func NewConfigurationManager(dataManager data.DataManager, networkManager network.NetworkManager,
	eventManager events.EventManager, sseClient realtime.SseClient, pollingUpdateInterval time.Duration,
	environment string, readiness ClientReadiness,
) *configurationManagerImpl {
	return &configurationManagerImpl{
		dataManager:           dataManager,
		networkManager:        networkManager,
		eventManager:          eventManager,
		sseClient:             sseClient,
		pollingUpdateInterval: pollingUpdateInterval,
		environment:           environment,
		readiness:             readiness,
	}
}

func (cm *configurationManagerImpl) Start() error {
	logging.Debug("CALL: configurationManagerImpl.Start()")
	// The fallback to the polling mode on failure is handled by `TryFetch` itself.
	err := cm.TryFetch(-1)
	logging.Debug("RETURN: configurationManagerImpl.Start() -> (error: %s)", err)
	return err
}

func (cm *configurationManagerImpl) OnUpdateConfiguration(handler func()) {
	logging.Debug("CALL: configurationManagerImpl.OnUpdateConfiguration()")
	cm.updateConfigurationHandler = handler
	logging.Debug("RETURN: configurationManagerImpl.OnUpdateConfiguration()")
}

func (cm *configurationManagerImpl) markReady() {
	if cm.readiness != nil {
		cm.readiness.MarkReady()
	}
}

func (cm *configurationManagerImpl) markNotReady(cause error) {
	if cm.readiness != nil {
		cm.readiness.MarkNotReady(cause)
	}
}

func (cm *configurationManagerImpl) TryFetch(ts int64) error {
	logging.Debug("CALL: configurationManagerImpl.tryFetch(ts: %s)", ts)
	err := cm.fetchConfig(ts)
	if err != nil {
		logging.Error("Fetch failed: %s", err)
	}
	// A failed update always falls back to polling: a streaming event means the
	// configuration has changed, so staying in the streaming mode would leave the SDK
	// serving the outdated configuration until the next event arrives.
	cm.manageConfigurationUpdate((err == nil) && cm.dataManager.DataFile().Settings().RealTimeUpdate())
	if err != nil {
		cm.markNotReady(err)
		logging.Debug("RETURN: configurationManagerImpl.tryFetch(ts: %s) -> (error: %s)", ts, err)
		return err
	}
	cm.markReady()
	logging.Debug("RETURN: configurationManagerImpl.tryFetch(ts: %s) -> (error: <nil>)", ts)
	return nil
}

func (cm *configurationManagerImpl) fetchConfig(ts int64) error {
	logging.Debug("CALL: configurationManagerImpl.fetchConfig(ts: %s)", ts)
	clientConfig, hasClientConfig, lastModified, err := cm.requestClientConfig(ts)
	if err == nil {
		if hasClientConfig {
			source := events.DataFileUpdateSourcePolling
			if ts != -1 {
				source = events.DataFileUpdateSourceStreaming
			}
			cm.updateDataFile(NewDataFile(clientConfig, lastModified, cm.environment), source)
		}
	} else {
		logging.Error("Failed to fetch: %s", err)
	}
	logging.Debug("RETURN: configurationManagerImpl.fetchConfig(ts: %s) -> (error: %s)", ts, err)
	return err
}

func (cm *configurationManagerImpl) updateDataFile(df *DataFile, source events.DataFileUpdateSource) {
	logging.Debug("CALL: configurationManagerImpl.updateDataFile(df: %s, source: %s)", df, source)
	cm.mx.Lock()
	cm.dataManager.SetDataFile(df)
	cm.networkManager.GetUrlProvider().ApplyDataApiDomain(df.Settings().DataApiDomain())
	cm.mx.Unlock()
	cm.eventManager.FireDataFileUpdate(events.DataFileUpdateEvent{
		Source:       source,
		DateModified: df.DateModified(),
	})
	if source == events.DataFileUpdateSourceStreaming {
		cm.fireDeprecatedUpdateConfigurationHandler()
	}
	logging.Debug("RETURN: configurationManagerImpl.updateDataFile(df: %s, source: %s)", df, source)
}

// fireDeprecatedUpdateConfigurationHandler calls the deprecated update configuration handler
// when the configuration was updated by a real-time (streaming) notification.
func (cm *configurationManagerImpl) fireDeprecatedUpdateConfigurationHandler() {
	handler := cm.updateConfigurationHandler
	if handler == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			logging.Warning("Update configuration handler failed: %s", r)
		}
	}()
	handler()
}

func (cm *configurationManagerImpl) requestClientConfig(ts int64) (Configuration, bool, string, error) {
	logging.Debug("CALL: configurationManagerImpl.requestClientConfig(ts: %s)", ts)
	if ts == -1 {
		logging.Info("Fetching configuration")
	} else {
		logging.Info("Fetching configuration for TS: %s", ts)
	}
	var campaigns Configuration

	fetchedConfiguration, err := cm.networkManager.FetchConfiguration(ts, cm.dataManager.DataFile().LastModified())
	if (err == nil) && (len(fetchedConfiguration.Configuration) > 0) {
		err = json.Unmarshal(fetchedConfiguration.Configuration, &campaigns)
	}
	if err == nil {
		logging.Info("Configuraiton fetched: %s", campaigns)
	} else {
		logging.Error("Failed to fetch client-config: %s", err)
	}
	logging.Debug("RETURN: configurationManagerImpl.requestClientConfig(ts: %s) -> (campaigns: %s, error: %s)",
		ts, campaigns, err)
	return campaigns, len(fetchedConfiguration.Configuration) > 0, fetchedConfiguration.LastModified, err
}

// manageConfigurationUpdate switches the SDK between the streaming and the polling
// configuration update mode. All switches are serialized by `updateModeMx`, so concurrent
// fetch outcomes cannot leave the SDK with duplicated or missing update loops.
func (cm *configurationManagerImpl) manageConfigurationUpdate(realTimeMode bool) {
	logging.Debug("CALL: configurationManagerImpl.manageConfigurationUpdate(realTimeMode: %s)", realTimeMode)
	cm.updateModeMx.Lock()
	defer cm.updateModeMx.Unlock()
	if realTimeMode {
		cm.switchToStreamingMode()
	} else {
		cm.switchToPollingMode()
	}
	logging.Debug("RETURN: configurationManagerImpl.manageConfigurationUpdate(realTimeMode: %s)", realTimeMode)
}

// switchToStreamingMode starts the streaming service if it is not already running. The
// polling loop is not stopped here: it notices the running streaming service on its next
// iteration and exits itself (see `runPollingLoop`). Must be called under `updateModeMx`.
func (cm *configurationManagerImpl) switchToStreamingMode() {
	if cm.realTimeEventService != nil {
		return
	}
	logging.Debug("CALL: configurationManagerImpl.switchToStreamingMode()")
	updateChan := make(chan realtime.RealTimeEvent, 16)
	cm.realTimeUpdateChan = updateChan
	cm.realTimeEventService = realtime.NewRealTimeEventService(
		cm.networkManager.GetUrlProvider().MakeRealTimeUrl(), updateChan, cm.sseClient)
	go func() {
		for realTimeEvent := range updateChan {
			cm.TryFetch(realTimeEvent.TimeStamp)
		}
	}()
	logging.Info("Configuration streaming is started")
	logging.Debug("RETURN: configurationManagerImpl.switchToStreamingMode()")
}

// switchToPollingMode stops the streaming service and ensures a polling loop is running.
// Must be called under `updateModeMx`.
func (cm *configurationManagerImpl) switchToPollingMode() {
	if cm.realTimeEventService != nil {
		cm.realTimeEventService.Close()
		cm.realTimeEventService = nil
		logging.Info("Configuration streaming is stopped")
	}
	if cm.pollingRunning {
		return // a single polling loop owns the schedule
	}
	logging.Debug("CALL: configurationManagerImpl.switchToPollingMode()")
	cm.pollingRunning = true
	go cm.runPollingLoop()
	logging.Info("Configuration polling is started")
	logging.Debug("RETURN: configurationManagerImpl.switchToPollingMode()")
}

// runPollingLoop fetches the configuration every `pollingUpdateInterval` until the SDK
// switches to the streaming mode. The loop terminates itself rather than being stopped
// from the outside: it may be the very goroutine whose fetch triggers the switch to
// streaming, so an external synchronous stop would deadlock.
func (cm *configurationManagerImpl) runPollingLoop() {
	for {
		time.Sleep(cm.pollingUpdateInterval)
		cm.updateModeMx.Lock()
		if (cm.realTimeEventService != nil) && cm.realTimeEventService.IsConnected() {
			cm.pollingRunning = false
			cm.updateModeMx.Unlock()
			logging.Info("Configuration polling is stopped")
			return
		}
		cm.updateModeMx.Unlock()
		cm.TryFetch(-1)
	}
}

func (cm *configurationManagerImpl) isPollingRunning() bool {
	cm.updateModeMx.Lock()
	defer cm.updateModeMx.Unlock()
	return cm.pollingRunning
}
