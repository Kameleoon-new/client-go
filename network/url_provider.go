package network

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
)

const (
	TestDataApiDomain          = "data.kameleoon.net"
	DefaultDataApiDomain       = "data.kameleoon.io"
	DefaultEventsDomain        = "events.kameleoon.eu"
	DefaultConfigurationDomain = "sdk-config.kameleoon.eu"
	DefaultAccessTokenDomain   = "api.kameleoon.com"

	trackingPath    = "/visit/events"
	visitorDataPath = "/visit/visitor"
	getDataPath     = "/map/map"

	configurationApiUrlFormat = "https://%s/v3/%s"
	rtConfigurationUrlFormat  = "https://%s:8110/sse?%s"
	accessTokenUrlFormat      = "https://%s/oauth/token?%s"
	dataApiUrlFormat          = "https://%s%s?%s"
)

type UrlProvider interface {
	MakeTrackingUrl() string
	MakeVisitorDataGetUrl(visitorCode string, filter types.RemoteVisitorDataFilter, isUniqueIdentifier bool) string
	MakeApiDataGetRequestUrl(key string) string
	MakeConfigurationUrl(environment string, ts int64) string
	MakeRealTimeUrl() string
	MakeAccessTokenUrl() string

	ApplyDataApiDomain(dataApiDomain string)

	SiteCode() string
	DataApiDomain() string
	EventsDomain() string
	ConfigurationDomain() string
	AccessTokenDomain() string
	SdkName() string
	SdkVersion() string
}

type UrlProviderImpl struct {
	siteCode             string
	dataApiDomain        string
	eventsDomain         string
	configurationDomain  string
	accessTokenDomain    string
	sdkName              string
	sdkVersion           string
	isCustomDomain       bool
	accessTokenQueryBase string
	trackingQueryBase    string
}

func NewUrlProviderImpl(siteCode string, networkDomain string, sdkName string, sdkVersion string) *UrlProviderImpl {
	queryBuilder := makeSdkSiteCodeQueryBase(siteCode, sdkName, sdkVersion)
	accessTokenQueryBase := queryBuilder.String()
	queryBuilder.Append(utils.QPBodyUA, "true")
	trackingQueryBase := queryBuilder.String()

	up := &UrlProviderImpl{
		siteCode:             siteCode,
		sdkName:              sdkName,
		sdkVersion:           sdkVersion,
		dataApiDomain:        DefaultDataApiDomain,
		eventsDomain:         DefaultEventsDomain,
		configurationDomain:  DefaultConfigurationDomain,
		accessTokenDomain:    DefaultAccessTokenDomain,
		accessTokenQueryBase: accessTokenQueryBase,
		trackingQueryBase:    trackingQueryBase,
	}
	up.updateDomains(networkDomain)
	return up
}

func makeSdkSiteCodeQueryBase(siteCode string, sdkName string, sdkVersion string) *utils.QueryBuilder {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPSdkName, sdkName)
	qb.Append(utils.QPSdkVersion, sdkVersion)
	qb.Append(utils.QPSiteCode, siteCode)
	return qb
}

func (up *UrlProviderImpl) SiteCode() string {
	return up.siteCode
}

func (up *UrlProviderImpl) DataApiDomain() string {
	return up.dataApiDomain
}

func (up *UrlProviderImpl) EventsDomain() string {
	return up.eventsDomain
}

func (up *UrlProviderImpl) ConfigurationDomain() string {
	return up.configurationDomain
}

func (up *UrlProviderImpl) AccessTokenDomain() string {
	return up.accessTokenDomain
}

func (up *UrlProviderImpl) SdkName() string {
	return up.sdkName
}

func (up *UrlProviderImpl) SdkVersion() string {
	return up.sdkVersion
}

func getUserIdQP(isUniqueIdentifier bool) string {
	if isUniqueIdentifier {
		return utils.QPMappingValue
	}
	return utils.QPVisitorCode
}

func (up *UrlProviderImpl) updateDomains(networkDomain string) {
	if networkDomain == "" {
		return
	}
	up.isCustomDomain = true

	up.eventsDomain = "events." + networkDomain
	up.configurationDomain = "sdk-config." + networkDomain
	up.dataApiDomain = "data." + networkDomain
	up.accessTokenDomain = "api." + networkDomain
}

func (up *UrlProviderImpl) ApplyDataApiDomain(dataApiDomain string) {
	if dataApiDomain != "" {
		if up.isCustomDomain {
			subDomain := dataApiDomain[:strings.Index(dataApiDomain, ".")]
			re := regexp.MustCompile("^[^.]+")
			up.dataApiDomain = re.ReplaceAllString(up.dataApiDomain, subDomain)
		} else {
			up.dataApiDomain = dataApiDomain
		}
	}
}

func (up *UrlProviderImpl) MakeTrackingUrl() string {
	return fmt.Sprintf(dataApiUrlFormat, up.dataApiDomain, trackingPath, up.trackingQueryBase)
}

func (up *UrlProviderImpl) MakeVisitorDataGetUrl(
	visitorCode string,
	filter types.RemoteVisitorDataFilter,
	isUniqueIdentifier bool,
) string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPSiteCode, up.siteCode)
	qb.Append(getUserIdQP(isUniqueIdentifier), visitorCode)
	qb.Append(utils.QPMaxNumberPreviousVisits, fmt.Sprintf("%d", filter.PreviousVisitAmount))
	qb.Append(utils.QPVersion, "0")
	qb.Append(utils.QPStaticData, "true")
	addFlagParamIfRequired(qb, utils.QPKcs, filter.Kcs)
	addFlagParamIfRequired(qb, utils.QPCurrentVisit, filter.CurrentVisit)
	addFlagParamIfRequired(qb, utils.QPCustomData, filter.CustomData)
	addFlagParamIfRequired(qb, utils.QPConversion, filter.Conversion)
	addFlagParamIfRequired(qb, utils.QPGeolocation, filter.Geolocation)
	addFlagParamIfRequired(qb, utils.QPExperiment, filter.Experiments)
	addFlagParamIfRequired(qb, utils.QPPage, filter.PageViews)
	addFlagParamIfRequired(qb, utils.QPPersonalization, filter.Personalization)
	addFlagParamIfRequired(qb, utils.QPCbs, filter.Cbs)
	return fmt.Sprintf(dataApiUrlFormat, up.dataApiDomain, visitorDataPath, qb.String())
}

func addFlagParamIfRequired(qb *utils.QueryBuilder, paramName string, state bool) {
	if state {
		qb.Append(paramName, "true")
	}
}

func (up *UrlProviderImpl) MakeApiDataGetRequestUrl(key string) string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPSiteCode, up.siteCode)
	qb.Append(utils.QPKey, key)
	return fmt.Sprintf(dataApiUrlFormat, up.dataApiDomain, getDataPath, qb.String())
}

func (up *UrlProviderImpl) MakeConfigurationUrl(environment string, ts int64) string {
	type param struct {
		name, value string
	}
	params := make([]param, 0, 2)
	if len(environment) > 0 {
		params = append(params, param{name: utils.QPEnvironment, value: environment})
	}
	if ts != -1 {
		params = append(params, param{name: utils.QPTimestamp, value: fmt.Sprint(ts)})
	}
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(configurationApiUrlFormat, up.configurationDomain, up.siteCode))
	for i := 0; i < len(params); i++ {
		if i == 0 {
			sb.WriteRune('?')
		} else {
			sb.WriteRune('&')
		}
		sb.WriteString(params[i].name)
		sb.WriteRune('=')
		sb.WriteString(params[i].value)
	}
	return sb.String()
}

func (up *UrlProviderImpl) MakeRealTimeUrl() string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPSiteCode, up.siteCode)
	return fmt.Sprintf(rtConfigurationUrlFormat, up.eventsDomain, qb)
}

func (up *UrlProviderImpl) MakeAccessTokenUrl() string {
	return fmt.Sprintf(accessTokenUrlFormat, up.accessTokenDomain, up.accessTokenDomain)
}
