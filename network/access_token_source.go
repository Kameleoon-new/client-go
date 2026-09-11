package network

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/utils"
)

const (
	SilencePeriod            = 5 * time.Minute
	TokenExpirationGap       = 1 * time.Minute
	TokenObsolescenceGap     = 30 * time.Minute
	BasicAuthorizationPrefix = "Basic "
)

type AccessTokenSource interface {
	GetToken(timeout time.Duration) string
	DiscardToken(token string)
}

type AccessTokenSourceImpl struct {
	clientId                string
	clientSecret            string
	networkManager          NetworkManager
	basicAuthorizationToken string

	mx                           sync.Mutex // guards the fields below
	cachedToken                  *expiringToken
	silentAfterFetchFailureUntil time.Time   // zero = never silenced
	fetching                     *tokenFetch // in-flight fetch shared by all concurrent callers; nil when idle
}

// tokenFetch is the result of one fetch request; `token` is readable once `done` is closed.
type tokenFetch struct {
	done  chan struct{}
	token string
}

func NewAccessTokenSource(clientId string, clientSecret string, networkManager NetworkManager) *AccessTokenSourceImpl {
	return &AccessTokenSourceImpl{
		networkManager:          networkManager,
		clientId:                clientId,
		clientSecret:            clientSecret,
		basicAuthorizationToken: constructBasicToken(clientId, clientSecret),
	}
}

func constructBasicToken(clientId string, clientSecret string) string {
	basicTokenContent := clientId + ":" + clientSecret
	return BasicAuthorizationPrefix + base64.StdEncoding.EncodeToString([]byte(basicTokenContent))
}

func (ats *AccessTokenSourceImpl) GetToken(timeout time.Duration) string {
	logging.Debug("CALL: AccessTokenSourceImpl.GetToken(timeout: %s)", timeout)
	now := time.Now()
	var resultToken string
	var awaited *tokenFetch
	ats.mx.Lock()
	token := ats.cachedToken
	silent := now.Before(ats.silentAfterFetchFailureUntil)
	if token != nil && !token.isExpired(now) {
		if token.isObsolete(now) && !silent {
			ats.fetchToken(timeout) // refresh in background, keep serving the current token
		}
		resultToken = token.value
	} else if !silent {
		awaited = ats.fetchToken(timeout)
	}
	ats.mx.Unlock()
	if awaited != nil {
		resultToken = awaited.wait(timeout)
	}
	logging.Debug("RETURN: AccessTokenSourceImpl.GetToken(timeout: %s) -> (token: '%s')", timeout,
		utils.Secret(resultToken))
	return resultToken
}

// Waits for the fetch up to `timeout` (unbounded when non-positive) and returns its token, or "" when the caller's
// own timeout expires first; the fetch itself carries on for the other waiters and the cache.
func (f *tokenFetch) wait(timeout time.Duration) string {
	if timeout <= 0 {
		<-f.done
		return f.token
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-f.done:
		return f.token
	case <-timer.C:
		logging.Warning("Access token is not fetched within %s; the request proceeds without it", timeout)
		return ""
	}
}

func (ats *AccessTokenSourceImpl) DiscardToken(token string) {
	logging.Debug("CALL: AccessTokenSourceImpl.DiscardToken(token: '%s')", utils.Secret(token))
	ats.mx.Lock()
	if ats.cachedToken != nil && ats.cachedToken.value == token {
		ats.cachedToken = nil
	}
	ats.mx.Unlock()
	logging.Debug("RETURN: AccessTokenSourceImpl.DiscardToken(token: '%s')", utils.Secret(token))
}

// Starts a fetch unless one is already in flight; returns the fetch shared by all waiters. Requires `mx` held.
func (ats *AccessTokenSourceImpl) fetchToken(timeout time.Duration) *tokenFetch {
	if ats.fetching == nil {
		ats.fetching = &tokenFetch{done: make(chan struct{})}
		go ats.performFetch(ats.fetching, timeout)
	}
	return ats.fetching
}

func (ats *AccessTokenSourceImpl) performFetch(fetch *tokenFetch, timeout time.Duration) {
	logging.Debug("CALL: AccessTokenSourceImpl.performFetch(timeout: %s)", timeout)
	response, err := ats.requestToken(timeout)
	ats.mx.Lock()
	if err == nil {
		ats.cachedToken = newExpiringToken(response.Token, response.ExpiresIn)
		fetch.token = response.Token
		logging.Info("Fetched access token")
	} else {
		ats.silentAfterFetchFailureUntil = time.Now().Add(SilencePeriod)
		logging.Warning("Failed to fetch access token (%s); it will not be requested for %s", err, SilencePeriod)
	}
	// Clear before completing so waiters that re-enter GetToken() can start a new fetch.
	ats.fetching = nil
	ats.mx.Unlock()
	close(fetch.done)
	logging.Debug("RETURN: AccessTokenSourceImpl.performFetch(timeout: %s) -> (token: '%s')", timeout,
		utils.Secret(fetch.token))
}

func (ats *AccessTokenSourceImpl) requestToken(timeout time.Duration) (accessTokenResponse, error) {
	response := accessTokenResponse{}
	jsonResponse, err := ats.networkManager.FetchAccessJWToken(ats.basicAuthorizationToken, timeout)
	if err != nil {
		return response, err
	}
	if err = json.Unmarshal(jsonResponse, &response); err != nil {
		return response, err
	}
	if response.Token == "" {
		return response, errors.New("access token JSON response has no 'access_token' field")
	}
	return response, nil
}

func newExpiringToken(token string, expiresInSeconds int) *expiringToken {
	expiresIn := time.Duration(expiresInSeconds) * time.Second
	now := time.Now()
	expTime := now.Add(expiresIn - TokenExpirationGap)
	var obsTime time.Time
	if expiresIn > TokenObsolescenceGap {
		obsTime = now.Add(expiresIn - TokenObsolescenceGap)
	} else {
		obsTime = expTime
		if expiresIn <= TokenExpirationGap {
			logging.Error("Access token life time (%ss) is not long enough to cache the token", expiresInSeconds)
		} else {
			logging.Warning("Access token life time (%ss) is not long enough to refresh cached token in background",
				expiresInSeconds)
		}
	}
	return &expiringToken{value: token, expirationTime: expTime, obsolescenceTime: obsTime}
}

type expiringToken struct {
	value            string
	expirationTime   time.Time
	obsolescenceTime time.Time
}

func (et *expiringToken) isExpired(now time.Time) bool {
	return !now.Before(et.expirationTime)
}

func (et *expiringToken) isObsolete(now time.Time) bool {
	return !now.Before(et.obsolescenceTime)
}

type accessTokenResponse struct {
	Token     string `json:"access_token"`
	ExpiresIn int    `json:"expires_in"`
}
