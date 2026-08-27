package network

import (
	"encoding/json"
	"time"

	"github.com/Kameleoon/client-go/v3/events"
	"github.com/Kameleoon/client-go/v3/types"
)

const (
	DefaultTrackingCallRetryDelay = time.Second * 5
)

func (nm *NetworkManagerImpl) GetRemoteData(key string, timeout time.Duration) (json.RawMessage, error) {
	url := nm.UrlProvider.MakeApiDataGetRequestUrl(key)
	request := Request{
		Method:         HttpGet,
		Url:            url,
		ContentType:    JsonContentType,
		Timeout:        timeout,
		IsAuthRequired: true,
	}
	response, err := nm.makeCall(events.RequestTypeRemoteData, &request, NetworkCallAttemptsNumberUncritical, -1)
	return response.Body, err
}

func (nm *NetworkManagerImpl) GetRemoteVisitorData(
	visitorCode string, filter types.RemoteVisitorDataFilter, isUniqueIdentifier bool, timeout time.Duration,
) (json.RawMessage, error) {
	url := nm.UrlProvider.MakeVisitorDataGetUrl(visitorCode, filter, isUniqueIdentifier)
	request := Request{
		Method:         HttpGet,
		Url:            url,
		ContentType:    JsonContentType,
		Timeout:        timeout,
		IsAuthRequired: true,
	}
	response, err := nm.makeCall(
		events.RequestTypeRemoteVisitorData, &request, NetworkCallAttemptsNumberUncritical, -1)
	return response.Body, err
}

func (nm *NetworkManagerImpl) SendTrackingData(trackingLines string) (bool, error) {
	if trackingLines == "" {
		return false, nil
	}
	url := nm.UrlProvider.MakeTrackingUrl()
	request := Request{
		Method:         HttpPost,
		Url:            url,
		ContentType:    WildcardContentType,
		Data:           trackingLines,
		Timeout:        nm.DefaultTimeout,
		IsAuthRequired: true,
	}
	_, err := nm.makeCall(
		events.RequestTypeTracking, &request, NetworkCallAttemptsNumberCritical, nm.TrackingCallRetryDelay)
	if err != nil {
		return false, err
	}
	return true, nil
}
