package configuration

import (
	"sync"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"

	"github.com/Kameleoon/client-go/v3/events"
	"github.com/Kameleoon/client-go/v3/managers/data"
	"github.com/Kameleoon/client-go/v3/network"
	"github.com/Kameleoon/client-go/v3/network/streaming"
	"github.com/segmentio/encoding/json"
)

// Not thread-safe
type ConfigurationManager interface {
	Start() error
	OnUpdateConfiguration(handler func())
	TryFetch(ts int64) error
	// Close stops the configuration updates (polling and streaming) for good.
	Close()
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
	sseClient      streaming.SseClient

	pollingUpdateInterval time.Duration
	streamingRetryDelay   time.Duration
	environment           string

	updateConfigurationHandler func()
	readiness                  ClientReadiness

	// mx guards the comparison and the replacement of the data file in `updateDataFile`; it is
	// never held across I/O.
	mx sync.Mutex

	// updateModeMx serializes the switches between the streaming and the polling
	// configuration update mode and guards the fields below, so concurrent fetch
	// outcomes cannot interleave the mode transitions.
	updateModeMx sync.Mutex
	// pollingStop belongs to the running polling loop, which exits once it is closed; nil while
	// no loop is running.
	pollingStop      chan struct{}
	streamingService *streaming.StreamingService
	closed           bool
}

// streamingRetryDelay is the delay before the first poll after streaming became unavailable
// or a fetch failed while the settings enable streaming; the poll re-enables streaming if it
// is still wanted.
const streamingRetryDelay = time.Minute

func NewConfigurationManager(dataManager data.DataManager, networkManager network.NetworkManager,
	eventManager events.EventManager, sseClient streaming.SseClient, pollingUpdateInterval time.Duration,
	environment string, readiness ClientReadiness,
) *configurationManagerImpl {
	return &configurationManagerImpl{
		dataManager:           dataManager,
		networkManager:        networkManager,
		eventManager:          eventManager,
		sseClient:             sseClient,
		pollingUpdateInterval: pollingUpdateInterval,
		streamingRetryDelay:   streamingRetryDelay,
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

// Close stops the polling loop and the streaming service. A fetch which is in flight cannot
// restart them: `manageConfigurationUpdate` does nothing once the manager is closed.
func (cm *configurationManagerImpl) Close() {
	logging.Debug("CALL: configurationManagerImpl.Close()")
	cm.updateModeMx.Lock()
	cm.closed = true
	cm.stopStreaming()
	cm.stopPolling()
	cm.updateModeMx.Unlock()
	logging.Debug("RETURN: configurationManagerImpl.Close()")
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
	cm.manageConfigurationUpdate(err == nil, cm.dataManager.DataFile().Settings().StreamingMode())
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

// updateDataFile replaces the data file with `df` unless it is older than the current one.
// Fetches may run concurrently (a streaming event can arrive while a poll is in flight) and
// finish out of order, and a stale reply must not roll the configuration back. The comparison
// and the replacement are atomic, and the update events fire under the same lock, so handlers
// see the data files in the applied order.
func (cm *configurationManagerImpl) updateDataFile(df *DataFile, source events.DataFileUpdateSource) {
	logging.Debug("CALL: configurationManagerImpl.updateDataFile(df: %s, source: %s)", df, source)
	cm.mx.Lock()
	currentDateModified := cm.dataManager.DataFile().DateModified()
	if df.DateModified() < currentDateModified {
		logging.Info("Fetched configuration (dateModified: %s) is older than the current one (dateModified: %s), "+
			"skipping", df.DateModified(), currentDateModified)
	} else {
		cm.dataManager.SetDataFile(df)
		cm.networkManager.GetUrlProvider().ApplyDataApiDomain(df.Settings().DataApiDomain())
		cm.eventManager.FireDataFileUpdate(events.DataFileUpdateEvent{
			Source:       source,
			DateModified: df.DateModified(),
		})
		if source == events.DataFileUpdateSourceStreaming {
			cm.fireDeprecatedUpdateConfigurationHandler()
		}
	}
	cm.mx.Unlock()
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
// configuration update mode once a fetch has settled. A failed fetch always falls back to
// polling: a streaming event means the configuration has changed, so staying in the streaming
// mode would leave the SDK serving the outdated configuration until the next event arrives.
// `streamingMode` tells whether the current settings enable streaming. All switches are
// serialized by `updateModeMx`, so concurrent fetch outcomes cannot leave the SDK with
// duplicated or missing update loops.
func (cm *configurationManagerImpl) manageConfigurationUpdate(fetchSucceeded bool, streamingMode bool) {
	logging.Debug("CALL: configurationManagerImpl.manageConfigurationUpdate(fetchSucceeded: %s, streamingMode: %s)",
		fetchSucceeded, streamingMode)
	cm.updateModeMx.Lock()
	defer cm.updateModeMx.Unlock()
	if cm.closed {
		logging.Debug("Configuration updates are not scheduled: the configuration manager is closed")
	} else if fetchSucceeded && streamingMode {
		cm.switchToStreamingMode()
	} else {
		cm.switchToPollingMode(cm.firstPollDelay(streamingMode))
	}
	logging.Debug("RETURN: configurationManagerImpl.manageConfigurationUpdate(fetchSucceeded: %s, streamingMode: %s)",
		fetchSucceeded, streamingMode)
}

// firstPollDelay is the delay before the first poll when the SDK falls back to polling. While
// the settings enable streaming, it is short enough to retry streaming soon (each successful
// poll switches back to it) and never longer than the polling interval; otherwise, it is the
// polling interval.
func (cm *configurationManagerImpl) firstPollDelay(streamingMode bool) time.Duration {
	if streamingMode && (cm.streamingRetryDelay < cm.pollingUpdateInterval) {
		return cm.streamingRetryDelay
	}
	return cm.pollingUpdateInterval
}

// switchToStreamingMode starts the streaming service if it is not already running, and stops
// the polling loop. Must be called under `updateModeMx`.
func (cm *configurationManagerImpl) switchToStreamingMode() {
	logging.Debug("CALL: configurationManagerImpl.switchToStreamingMode()")
	if cm.streamingService == nil {
		cm.streamingService = streaming.NewStreamingService(
			cm.networkManager.GetUrlProvider().MakeRealTimeUrl(), cm.sseClient, cm.onDataFileUpdateMessage,
			cm.onStreamingUnavailable)
		logging.Info("Configuration streaming is started")
	}
	cm.stopPolling()
	logging.Debug("RETURN: configurationManagerImpl.switchToStreamingMode()")
}

// switchToPollingMode stops the streaming service and ensures a polling loop is running.
// `initialDelay` is the delay before the first poll; the following ones are
// `pollingUpdateInterval` apart. It is ignored while a polling loop is already running, as
// a single loop owns the schedule. Must be called under `updateModeMx`.
func (cm *configurationManagerImpl) switchToPollingMode(initialDelay time.Duration) {
	logging.Debug("CALL: configurationManagerImpl.switchToPollingMode(initialDelay: %s)", initialDelay)
	cm.stopStreaming()
	if cm.pollingStop == nil {
		stop := make(chan struct{})
		cm.pollingStop = stop
		go cm.runPollingLoop(stop, initialDelay)
		logging.Info("Configuration polling is started")
	}
	logging.Debug("RETURN: configurationManagerImpl.switchToPollingMode(initialDelay: %s)", initialDelay)
}

// stopStreaming stops the streaming service, if any. Must be called under `updateModeMx`.
func (cm *configurationManagerImpl) stopStreaming() {
	if cm.streamingService != nil {
		cm.streamingService.Close()
		cm.streamingService = nil
		logging.Info("Configuration streaming is stopped")
	}
}

// stopPolling stops the running polling loop, if any, without waiting for it: the loop may be
// the very goroutine whose fetch triggers the switch to streaming, so waiting for it would
// deadlock. Must be called under `updateModeMx`.
func (cm *configurationManagerImpl) stopPolling() {
	if cm.pollingStop != nil {
		close(cm.pollingStop)
		cm.pollingStop = nil
		logging.Info("Configuration polling is stopped")
	}
}

// onDataFileUpdateMessage fetches the configuration a streaming event refers to, unless the
// current one is already as recent: the server sends the current configuration's timestamp
// on every (re)connection, and an update may have been fetched by a poll meanwhile.
func (cm *configurationManagerImpl) onDataFileUpdateMessage(message streaming.DataFileUpdateMessage) {
	dateModified := cm.dataManager.DataFile().DateModified()
	if message.TimeStamp <= dateModified {
		logging.Debug("Configuration is up to date (ts: %s, dateModified: %s), skipping update",
			message.TimeStamp, dateModified)
		return
	}
	cm.TryFetch(message.TimeStamp)
}

// onStreamingUnavailable falls back to polling once the streaming service has given up
// reconnecting. The first poll comes soon, so streaming is retried without waiting a whole
// polling interval; each successful poll re-enables it as long as the configuration asks for it.
func (cm *configurationManagerImpl) onStreamingUnavailable(service *streaming.StreamingService) {
	cm.updateModeMx.Lock()
	defer cm.updateModeMx.Unlock()
	if cm.streamingService != service {
		return // a replaced service reporting late, or the manager is closed
	}
	cm.streamingService = nil
	cm.switchToPollingMode(cm.firstPollDelay(true))
}

// runPollingLoop fetches the configuration after `initialDelay`, then every
// `pollingUpdateInterval`, until `stop` is closed.
func (cm *configurationManagerImpl) runPollingLoop(stop chan struct{}, initialDelay time.Duration) {
	timer := time.NewTimer(initialDelay)
	defer timer.Stop()
	for {
		select {
		case <-stop:
			return
		case <-timer.C:
		}
		select {
		case <-stop:
			return // stopped at the very moment the timer fired
		default:
		}
		cm.TryFetch(-1)
		timer.Reset(cm.pollingUpdateInterval)
	}
}
