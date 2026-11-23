package types

type TargetingType string

const (
	TargetingPageUrl                TargetingType = "PAGE_URL"
	TargetingPageTitle              TargetingType = "PAGE_TITLE"
	TargetingLandingPage            TargetingType = "LANDING_PAGE"
	TargetingOrigin                 TargetingType = "ORIGIN"
	TargetingOriginType             TargetingType = "ORIGIN_TYPE"
	TargetingReferrers              TargetingType = "REFERRERS"
	TargetingNewVisitors            TargetingType = "NEW_VISITORS"
	TargetingInterests              TargetingType = "INTERESTS"
	TargetingBrowserLanguage        TargetingType = "BROWSER_LANGUAGE"
	TargetingGeolocation            TargetingType = "GEOLOCATION"
	TargetingDeviceType             TargetingType = "DEVICE_TYPE"
	TargetingScreenDimension        TargetingType = "SCREEN_DIMENSION"
	TargetingVisitorIp              TargetingType = "VISITOR_IP"
	TargetingAdBlocker              TargetingType = "AD_BLOCKER"
	TargetingPreviousPage           TargetingType = "PREVIOUS_PAGE"
	TargetingKeyPages               TargetingType = "KEY_PAGES"
	TargetingPageViews              TargetingType = "PAGE_VIEWS"
	TargetingFirstVisit             TargetingType = "FIRST_VISIT"
	TargetingLastVisit              TargetingType = "LAST_VISIT"
	TargetingActiveSession          TargetingType = "ACTIVE_SESSION"
	TargetingTimeSincePageLoad      TargetingType = "TIME_SINCE_PAGE_LOAD"
	TargetingSameDayVisits          TargetingType = "SAME_DAY_VISITS"
	TargetingVisits                 TargetingType = "VISITS"
	TargetingVisitsByPage           TargetingType = "VISITS_BY_PAGE"
	TargetingInternalSearchKeywords TargetingType = "INTERNAL_SEARCH_KEYWORDS"
	TargetingTabsOnSite             TargetingType = "TABS_ON_SITE"
	TargetingConversionProbability  TargetingType = "CONVERSION_PROBABILITY"
	TargetingHeatSlice              TargetingType = "HEAT_SLICE"
	TargetingSkyStatus              TargetingType = "SKY_STATUS"
	TargetingTemperature            TargetingType = "TEMPERATURE"
	TargetingDayNight               TargetingType = "DAY_NIGHT"
	TargetingForecastSkyStatus      TargetingType = "FORECAST_SKY_STATUS"
	TargetingForecastTemperature    TargetingType = "FORECAST_TEMPERATURE"
	TargetingDayOfWeek              TargetingType = "DAY_OF_WEEK"
	TargetingTimeRange              TargetingType = "TIME_RANGE"
	TargetingHourMinuteRange        TargetingType = "HOUR_MINUTE_RANGE"
	TargetingJsCode                 TargetingType = "JS_CODE"
	TargetingCookie                 TargetingType = "COOKIE"
	TargetingEvent                  TargetingType = "EVENT"
	TargetingBrowser                TargetingType = "BROWSER"
	TargetingOperatingSystem        TargetingType = "OPERATING_SYSTEM"
	TargetingDomElement             TargetingType = "DOM_ELEMENT"
	TargetingMouseOut               TargetingType = "MOUSE_OUT"
	TargetingExperiments            TargetingType = "EXPERIMENTS"
	TargetingConversions            TargetingType = "CONVERSIONS"
	TargetingCustomDatum            TargetingType = "CUSTOM_DATUM"
	TargetingYsanceSegment          TargetingType = "YSANCE_SEGMENT"
	TargetingYsanceAttribut         TargetingType = "YSANCE_ATTRIBUT"
	TargetingTealiumBadge           TargetingType = "TEALIUM_BADGE"
	TargetingTealiumAudience        TargetingType = "TEALIUM_AUDIENCE"
	TargetingPriceOfDisplayedPage   TargetingType = "PRICE_OF_DISPLAYED_PAGE"
	TargetingNumberOfVisitedPages   TargetingType = "NUMBER_OF_VISITED_PAGES"
	TargetingVisitedPages           TargetingType = "VISITED_PAGES"
	TargetingMeanPageDuration       TargetingType = "MEAN_PAGE_DURATION"
	TargetingTimeSincePreviousVisit TargetingType = "TIME_SINCE_PREVIOUS_VISIT"
	TargetingTargetFeatureFlag      TargetingType = "TARGET_FEATURE_FLAG"
	TargetingTargetExperiment       TargetingType = "TARGET_EXPERIMENT"
	TargetingTargetPersonalization  TargetingType = "TARGET_PERSONALIZATION"
	TargetingExclusiveExperiment    TargetingType = "EXCLUSIVE_EXPERIMENT"
	TargetingSDKLanguage            TargetingType = "SDK_LANGUAGE"
	TargetingApplicationVersion     TargetingType = "APPLICATION_VERSION"
	TargetingVisitorCode            TargetingType = "VISITOR_CODE"
	TargetingSegment                TargetingType = "SEGMENT"
)

type OperatorType string

const (
	OperatorUndefined      OperatorType = "UNDEFINED"
	OperatorContains       OperatorType = "CONTAINS"
	OperatorExact          OperatorType = "EXACT"
	OperatorAny            OperatorType = "ANY"
	OperatorRegExp         OperatorType = "REGULAR_EXPRESSION"
	OperatorLower          OperatorType = "LOWER"
	OperatorEqual          OperatorType = "EQUAL"
	OperatorGreater        OperatorType = "GREATER"
	OperatorLowerOrEqual   OperatorType = "LOWER_OR_EQUAL"
	OperatorGreaterOrEqual OperatorType = "GREATER_OR_EQUAL"
	OperatorIsTrue         OperatorType = "TRUE"
	OperatorIsFalse        OperatorType = "FALSE"
	OperatorIsAmongValues  OperatorType = "AMONG_VALUES"
)

type TargetedDataSdk struct {
	Language string
	Version  string
}

type TargetingDataGetter func(TargetingType) interface{}
