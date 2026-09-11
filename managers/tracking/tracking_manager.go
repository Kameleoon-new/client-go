package tracking

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"

	"github.com/Kameleoon/client-go/v3/managers/data"
	"github.com/Kameleoon/client-go/v3/network"
	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/Kameleoon/client-go/v3/types"
)

type TrackingManager interface {
	AddVisitorCode(visitorCode string)

	TrackAll()
	TrackVisitor(visitorCode string)

	Close()
}

const (
	LinesDelimiter   = "\n"
	RequestSizeLimit = 2560 * 1024 // 2.5 * 1024^2 characters
	probeMaxVisitors = 1
)

type TrackingManagerImpl struct {
	trackingVisitors VisitorTrackingRegistry
	dataManager      data.DataManager
	networkManager   network.NetworkManager
	visitorManager   storage.VisitorManager
	trackingTicker   *time.Ticker
	stopChan         chan struct{}
	probeMode        int32 // atomic bool; set on request goroutines, read on the ticker goroutine
}

func NewTrackingManagerImpl(
	dataManager data.DataManager,
	networkManager network.NetworkManager,
	visitorManager storage.VisitorManager,
	trackInterval time.Duration,
) *TrackingManagerImpl {
	logging.Debug("CALL: NewTrackingManagerImpl(dataManager, networkManager, visitorManager, scheduledExecutor, "+
		"trackInterval: %s)", trackInterval)
	tm := &TrackingManagerImpl{
		trackingVisitors: NewConcurrentVisitorTrackingRegistry(
			visitorManager, DefaultStorageLimit, DefaultExtractionLimit,
		),
		dataManager:    dataManager,
		networkManager: networkManager,
		visitorManager: visitorManager,
		trackingTicker: time.NewTicker(trackInterval),
		stopChan:       make(chan struct{}, 8),
	}
	go func() {
		for {
			select {
			case <-tm.trackingTicker.C:
				tm.TrackAll()
			case <-tm.stopChan:
				return
			}
		}
	}()
	logging.Debug("RETURN: NewTrackingManagerImpl(dataManager, networkManager, visitorManager, scheduledExecutor, "+
		"trackInterval: %s) -> (TrackingManagerImpl)", trackInterval)
	return tm
}

func (tm *TrackingManagerImpl) Close() {
	logging.Debug("CALL: TrackingManagerImpl.Close()")
	tm.trackingTicker.Stop()
	if len(tm.stopChan) == 0 {
		tm.stopChan <- struct{}{}
	}
	logging.Debug("RETURN: TrackingManagerImpl.Close()")
}

func (tm *TrackingManagerImpl) AddVisitorCode(visitorCode string) {
	logging.Debug("CALL: TrackingManagerImpl.AddVisitorCode(visitorCode: %s)", visitorCode)
	tm.trackingVisitors.Add(visitorCode)
	logging.Debug("RETURN: TrackingManagerImpl.AddVisitorCode(visitorCode: %s)", visitorCode)
}

func (tm *TrackingManagerImpl) TrackAll() {
	logging.Debug("CALL: TrackingManagerImpl.TrackAll()")
	if tm.isProbeMode() {
		tm.trackProbe()
	} else {
		tm.track(tm.trackingVisitors.Extract(UnlimitedExtraction))
	}
	logging.Debug("RETURN: TrackingManagerImpl.TrackAll()")
}

func (tm *TrackingManagerImpl) TrackVisitor(visitorCode string) {
	logging.Debug("CALL: TrackingManagerImpl.TrackVisitor(visitorCode: %s)", visitorCode)
	tm.track([]string{visitorCode})
	logging.Debug("RETURN: TrackingManagerImpl.TrackVisitor(visitorCode: %s)", visitorCode)
}

// Sends one visitor's data; visitors without data are dropped (as in the normal path) until one is found.
func (tm *TrackingManagerImpl) trackProbe() {
	for {
		visitorCode := tm.trackingVisitors.Extract(probeMaxVisitors)
		if (len(visitorCode) == 0) || tm.track(visitorCode) {
			return
		}
	}
}

// Returns whether a request was sent.
func (tm *TrackingManagerImpl) track(visitorCodes []string) bool {
	builder := NewTrackingBuilder(visitorCodes, tm.dataManager.DataFile(), tm.visitorManager, RequestSizeLimit)
	builder.Build()
	if len(builder.VisitorCodesToKeep()) > 0 {
		logging.Warning(
			"Visitor data to be tracked exceeded the request size limit. " +
				"Some visitor data is kept to be sent later. " +
				"If it is not caused by the peak load, decreasing the tracking interval is recommended.",
		)
		tm.trackingVisitors.AddAll(builder.VisitorCodesToKeep())
	}
	return tm.performTrackingRequest(
		builder.VisitorCodesToSend(), builder.UnsentVisitorData(), builder.TrackingLines())
}

// Returns whether a request was sent.
func (tm *TrackingManagerImpl) performTrackingRequest(
	visitorCodes []string, unsentVisitorData []types.Sendable, trackingLines []string,
) bool {
	if len(trackingLines) == 0 {
		return false
	}
	// Mark unsent data as transmitted
	for _, s := range unsentVisitorData {
		s.MarkAsTransmitting()
	}
	lines := strings.Join(trackingLines, LinesDelimiter)
	go func() {
		out, err := tm.networkManager.SendTrackingData(lines)
		if (err == nil) && out {
			logging.Info("Successful request for tracking visitors: %s, data: %s", visitorCodes, unsentVisitorData)
			for _, s := range unsentVisitorData {
				s.MarkAsSent()
			}
			tm.leaveProbeMode()
		} else {
			logging.Error("Tracking request failed: %s", err)
			logging.Info("Failed request for tracking visitors: %s, data: %s", visitorCodes, unsentVisitorData)
			for _, s := range unsentVisitorData {
				s.MarkAsUnsent()
			}
			tm.trackingVisitors.AddAll(visitorCodes)
			tm.enterProbeMode()
		}
	}()
	return true
}

func (tm *TrackingManagerImpl) isProbeMode() bool {
	return atomic.LoadInt32(&tm.probeMode) != 0
}

func (tm *TrackingManagerImpl) enterProbeMode() {
	if atomic.CompareAndSwapInt32(&tm.probeMode, 0, 1) {
		logging.Info("Tracking request failed: only one visitor's data per tracking request will be sent " +
			"until a request succeeds")
	}
}

func (tm *TrackingManagerImpl) leaveProbeMode() {
	if atomic.CompareAndSwapInt32(&tm.probeMode, 1, 0) {
		logging.Info("Tracking request succeeded: full-size tracking requests are restored")
	}
}
