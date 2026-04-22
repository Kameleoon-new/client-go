package network

import (
	"encoding/json"
	"time"

	"github.com/Kameleoon/client-go/v3/errs"
	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/types"
)

const (
	networkCallRetriesNumberCritical    = 2
	NetworkCallAttemptsNumberCritical   = networkCallRetriesNumberCritical + 1 // +1 for initial request
	NetworkCallAttemptsNumberUncritical = 1

	codeUnauthorized = 401
)

// declaration

type NetworkManager interface {
	GetEnvironment() string
	GetDefaultTimeout() time.Duration
	GetNetProvider() NetProvider
	GetUrlProvider() UrlProvider
	GetAccessTokenSource() AccessTokenSource

	// Automation API
	FetchAccessJWToken(basicAuthorizationToken string, timeout time.Duration) (json.RawMessage, error)

	// SDK config API
	FetchConfiguration(ts int64, ifModifiedSince string) (FetchedConfiguration, error)

	// Data API
	GetRemoteData(key string, timeout time.Duration) (json.RawMessage, error)
	GetRemoteVisitorData(visitorCode string, filter types.RemoteVisitorDataFilter, isUniqueIdentifier bool,
		timeout time.Duration) (json.RawMessage, error)
	SendTrackingData(trackingLines string) (bool, error)
}

type FetchedConfiguration struct {
	Configuration json.RawMessage
	LastModified  string
}

// base implementation

type NetworkManagerImpl struct {
	Environment            string
	DefaultTimeout         time.Duration
	NetProvider            NetProvider
	UrlProvider            UrlProvider
	Logger                 logging.Logger
	TrackingCallRetryDelay time.Duration
	accessTokenSource      AccessTokenSource
}

func NewNetworkManagerImpl(
	environment string,
	defaultTimeout time.Duration,
	netProvider NetProvider,
	urlProvider UrlProvider,
	accessTokenSourceFactory AccessTokenSourceFactory,
) *NetworkManagerImpl {
	nm := &NetworkManagerImpl{
		Environment:            environment,
		DefaultTimeout:         defaultTimeout,
		NetProvider:            netProvider,
		UrlProvider:            urlProvider,
		TrackingCallRetryDelay: DefaultTrackingCallRetryDelay,
	}
	nm.accessTokenSource = accessTokenSourceFactory.create(nm)
	return nm
}

func (nm *NetworkManagerImpl) GetEnvironment() string {
	return nm.Environment
}

func (nm *NetworkManagerImpl) GetDefaultTimeout() time.Duration {
	return nm.DefaultTimeout
}

func (nm *NetworkManagerImpl) GetNetProvider() NetProvider {
	return nm.NetProvider
}

func (nm *NetworkManagerImpl) GetUrlProvider() UrlProvider {
	return nm.UrlProvider
}

func (nm *NetworkManagerImpl) GetAccessTokenSource() AccessTokenSource {
	return nm.accessTokenSource
}

// API call commons

func (nm *NetworkManagerImpl) ensureTimeout(request *Request) {
	if request.Timeout <= 0 {
		request.Timeout = nm.DefaultTimeout
	}
}

func (nm *NetworkManagerImpl) makeCall(
	request *Request, attemptCount int, retryDelay time.Duration, headersToRead ...string,
) (Response, error) {
	logging.Debug("Running request %s with retry limit %s, retry delay %s ms", request, attemptCount, retryDelay)
	nm.ensureTimeout(request)
	var err error
	var isTokenRejected bool
	var response Response
	for i := 0; i < attemptCount; i++ {
		logLevel := nm.getLogLevel(i, attemptCount)
		if (i > 0) && (retryDelay > 0) {
			time.Sleep(retryDelay)
		}
		nm.authorizeIfRequired(request)
		response = nm.NetProvider.Call(request, headersToRead)
		if isTokenRejected, err = nm.processErrors(request, &response, logLevel); err == nil {
			logging.Debug("Fetched response %s for request %s", response, request)
			return response, nil
		}
	}
	if isTokenRejected {
		logging.Error("Wrong Kameleoon API access token slows down the SDK's requests")
		request.Authorization = ""
		response = nm.NetProvider.Call(request, headersToRead)
		if _, err = nm.processErrors(request, &response, logging.ERROR); err == nil {
			logging.Debug("Fetched response %s for request %s", response, request)
			return response, nil
		}
	}
	return Response{}, err
}

func (nm *NetworkManagerImpl) getLogLevel(attempt int, attemptCount int) logging.LogLevel {
	if attempt == attemptCount-1 {
		return logging.ERROR
	}
	return logging.WARNING
}

func (nm *NetworkManagerImpl) authorizeIfRequired(request *Request) {
	if request.IsAuthRequired {
		request.Authorization = nm.accessTokenSource.GetToken(request.Timeout)
	}
}

func (nm *NetworkManagerImpl) processErrors(request *Request, response *Response, logLevel logging.LogLevel) (bool, error) {
	var err error
	var isTokenRejected bool
	if response.Err != nil {
		err = response.Err
		logging.Log(logLevel, "%s call %s failed: Error occurred during request: %s",
			request.Method, request.Url, err)
	} else if !response.IsExpectedStatusCode() {
		err = errs.NewUnexpectedStatusCode(response.Code, response.Body)
		logging.Log(logLevel, "%s call %s failed: Received unexpected status code: %s, body: %s",
			request.Method, request.Url, response.Code, string(response.Body[:]))
		if (response.Code == codeUnauthorized) && (request.Authorization != "") {
			logging.Log(logLevel, "Unexpected rejection of access token %s", request.Authorization)
			nm.accessTokenSource.DiscardToken(request.Authorization)
			isTokenRejected = true
		}
	}
	return isTokenRejected, err
}
