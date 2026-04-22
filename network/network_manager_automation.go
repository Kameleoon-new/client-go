package network

import (
	"encoding/json"
	"time"

	"github.com/Kameleoon/client-go/v3/utils"
)

const (
	GrantType             = "client_credentials"
	HeaderContentTypeName = "Content-Type"
)

func (nm *NetworkManagerImpl) FetchAccessJWToken(basicAuthorizationToken string,
	timeout time.Duration) (json.RawMessage, error) {

	url := nm.UrlProvider.MakeAccessTokenUrl()
	data := nm.formFetchAccessTokenData()
	request := Request{
		Method:        HttpPost,
		Url:           url,
		ContentType:   FormContentType,
		Authorization: basicAuthorizationToken,
		Data:          data,
		Timeout:       timeout,
	}
	response, err := nm.makeCall(&request, NetworkCallAttemptsNumberUncritical, -1)
	return response.Body, err
}

func (nm *NetworkManagerImpl) formFetchAccessTokenData() string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPGrantType, "client_credentials")
	return qb.String()
}
