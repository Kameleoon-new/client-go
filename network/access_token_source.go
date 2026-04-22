package network

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"
)

const (
	TokenExpirationGap       = 60   // in seconds
	TokenObsolescenceGap     = 1800 // in seconds
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
	cachedToken             *expiringToken // pointer for thread-safe
	fetching                bool
	basicAuthorizationToken string
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
	token := ats.cachedToken
	var resultToken string
	if token != nil && !token.isExpired(now) {
		if !ats.fetching && token.isObsolete(now) {
			ats.fetching = true // set `fetching` here as well to reduce the number of requests until goroutine runned
			go ats.fetchToken(timeout)
		}
		resultToken = token.value
	} else {
		resultToken = ats.fetchToken(timeout)
	}
	logging.Debug("RETURN: AccessTokenSourceImpl.GetToken(timeout: %s) -> (token: %s)", timeout, resultToken)
	return resultToken
}

func (ats *AccessTokenSourceImpl) DiscardToken(token string) {
	logging.Debug("CALL: AccessTokenSourceImpl.DiscardToken(token: %s)", token)
	cachedToken := ats.cachedToken
	if cachedToken != nil && cachedToken.value == token {
		ats.cachedToken = nil
	}
	logging.Debug("RETURN: AccessTokenSourceImpl.DiscardToken(token: %s)", token)
}

func (ats *AccessTokenSourceImpl) fetchToken(timeout time.Duration) string {
	logging.Debug("CALL: AccessTokenSourceImpl.fetchToken(timeout: %s)", timeout)
	ats.fetching = true
	defer func() { ats.fetching = false }()
	jsonResponse, err := ats.networkManager.FetchAccessJWToken(ats.basicAuthorizationToken, timeout)
	var token string
	if err != nil {
		logging.Error("Failed to read access JWT: %s", err)
		token = ""
	} else {
		accessTokenResponse := accessTokenResponse{}
		err = json.Unmarshal(jsonResponse, &accessTokenResponse)
		if err != nil {
			logging.Error("Failed to parse access JWT: %s", err)
			token = ""
		} else {
			ats.handleFetchedToken(accessTokenResponse)
			token = accessTokenResponse.Token
		}
	}
	logging.Debug("RETURN: AccessTokenSourceImpl.fetchToken(timeout: %s) -> (token: %s)", timeout, token)
	return token
}

func (ats *AccessTokenSourceImpl) handleFetchedToken(accessTokenResponse accessTokenResponse) {
	logging.Debug("CALL: AccessTokenSourceImpl.handleFetchedToken(accessTokenResponse: %s)", accessTokenResponse)
	expiresIn := accessTokenResponse.ExpiresIn
	now := time.Now()
	expTime := now.Add(time.Second * time.Duration(expiresIn-TokenExpirationGap))
	var obsTime time.Time
	if expiresIn > TokenObsolescenceGap {
		obsTime = now.Add(time.Second * time.Duration(expiresIn-TokenObsolescenceGap))
	} else {
		obsTime = expTime
		if expiresIn <= TokenExpirationGap {
			logging.Error("Access token life time (%ss) is not long enough to cache the token", expiresIn)
		} else {
			logging.Warning(
				"Access token life time (%ss) is not long enough to refresh cached token in background", expiresIn)
		}
	}
	ats.cachedToken = newExpiringToken(accessTokenResponse.Token, expTime, obsTime)
	logging.Debug("RETURN: AccessTokenSourceImpl.handleFetchedToken(accessTokenResponse: %s)", accessTokenResponse)
}

type expiringToken struct {
	value            string
	expirationTime   time.Time
	obsolescenceTime time.Time
}

func newExpiringToken(value string, expirationTime, obsolescenceTime time.Time) *expiringToken {
	return &expiringToken{
		value:            value,
		expirationTime:   expirationTime,
		obsolescenceTime: obsolescenceTime,
	}
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
