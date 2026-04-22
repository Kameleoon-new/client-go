package kameleoon

import (
	"fmt"
	"sync"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"

	"github.com/Kameleoon/client-go/v3/configuration"
	"github.com/Kameleoon/client-go/v3/errs"
	"github.com/Kameleoon/client-go/v3/managers/data"
	"github.com/Kameleoon/client-go/v3/managers/hybrid"
	"github.com/Kameleoon/client-go/v3/managers/remotedata"
	"github.com/Kameleoon/client-go/v3/managers/tracking"
	"github.com/Kameleoon/client-go/v3/managers/warehouse"
	"github.com/Kameleoon/client-go/v3/network"
	"github.com/Kameleoon/client-go/v3/network/cookie"
	"github.com/Kameleoon/client-go/v3/realtime"
	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/Kameleoon/client-go/v3/targeting"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
	"github.com/valyala/fasthttp"
)

// Deprecated: Please use `VisitorWarehouseAudienceOptParams` instead
type VisitorWarehouseAudienceParams struct {
	VisitorCode     string
	CustomDataIndex int
	WarehouseKey    string        // optional
	Timeout         time.Duration // optional
}

type VisitorWarehouseAudienceOptParams struct {
	WarehouseKey string
	Timeout      time.Duration
}

type RemoteVisitorDataOptParams struct {
	IsUniqueIdentifier bool // Deprecated: Please use `UniqueIdentifier` data instead
	Timeout            time.Duration
}

type GetVariationOptParams struct {
	track bool
}

func NewGetVariationOptParams() GetVariationOptParams {
	return GetVariationOptParams{track: true}
}
func (p GetVariationOptParams) Track(value bool) GetVariationOptParams {
	p.track = value
	return p
}

type GetVariationsOptParams struct {
	onlyActive bool
	track      bool
}

func NewGetVariationsOptParams() GetVariationsOptParams {
	return GetVariationsOptParams{onlyActive: false, track: true}
}
func (p GetVariationsOptParams) OnlyActive(value bool) GetVariationsOptParams {
	p.onlyActive = value
	return p
}
func (p GetVariationsOptParams) Track(value bool) GetVariationsOptParams {
	p.track = value
	return p
}

type AddDataOptParams struct {
	track bool
}

func NewAddDataOptParams() AddDataOptParams {
	return AddDataOptParams{track: true}
}

func (p AddDataOptParams) Track(value bool) AddDataOptParams {
	p.track = value
	return p
}

type TrackConversionOptParams struct {
	Revenue  float64
	Negative bool
	Metadata []*types.CustomData
}

type SetForcedVariationOptParams struct {
	forceTargeting bool
}

func NewSetForcedVariationOptParams() SetForcedVariationOptParams {
	return SetForcedVariationOptParams{forceTargeting: true}
}
func (p SetForcedVariationOptParams) ForceTargeting(value bool) SetForcedVariationOptParams {
	p.forceTargeting = value
	return p
}

type KameleoonClient interface {
	WaitInit() error

	// GetVisitorCode should be called to get the Kameleoon visitorCode for the current visitor.
	//
	// This is especially important when using Kameleoon in a mixed front-end and back-end environment,
	// where user identification consistency must be guaranteed.
	//
	// The implementation logic is described here:
	// First we check if a kameleoonVisitorCode cookie or query parameter associated with the current HTTP request can be
	// found. If so, we will use this as the visitor identifier. If no cookie / parameter is found in the current
	// request, we either randomly generate a new identifier, or use the defaultVisitorCode argument as identifier if it
	// is passed. This allows our customers to use their own identifiers as visitor codes, should they wish to.
	// This can have the added benefit of matching Kameleoon visitors with their own users without any additional
	// look-ups in a matching table.
	GetVisitorCode(request *fasthttp.Request, response *fasthttp.Response, defaultVisitorCode ...string) (string, error)

	SetLegalConsent(visitorCode string, consent bool, response ...*fasthttp.Response) error

	// AddData associates various Data with a visitor
	//
	// Note that this method doesn't return any value and doesn't interact with the
	// Kameleoon back-end servers by itself. Instead, the declared data is saved for
	// future sending via the Flush method. This reduces the number of server calls made,
	// as data is usually grouped into a single server call triggered by the execution
	// of the Flush method.
	AddData(visitorCode string, allData ...types.Data) error

	// AddDataWithOptParams associates various Data with a visitor, allowing additional
	// optional parameters to control how the data is handled.
	//
	// Note that this method doesn't return any value and doesn't interact with the
	// Kameleoon back-end servers by itself. Instead, the declared data is saved for
	// future sending via the Flush method. This reduces the number of server calls made,
	// as data is usually grouped into a single server call triggered by the execution
	// of the Flush method.
	AddDataWithOptParams(visitorCode string, optParams AddDataOptParams, allData ...types.Data) error

	// TrackConversion on a particular goal
	//
	// This method requires visitorCode and goalID to track conversion on this particular goal.
	// In addition, this method also accepts revenue as a third optional argument to track revenue.
	// This method is non-blocking as the server call is made asynchronously.
	TrackConversion(visitorCode string, goalID int, isUniqueIdentifier ...bool) error

	TrackConversionRevenue(visitorCode string, goalID int, revenue float64, isUniqueIdentifier ...bool) error

	TrackConversionWithOptParams(visitorCode string, goalId int, params TrackConversionOptParams) error

	// FlushVisitor the associated data.
	//
	// The data added with the method AddData, is not directly sent to the kameleoon servers.
	// It's stored and accumulated until it is sent automatically by the TriggerExperiment or TrackConversion methods.
	// With this method you can manually send it.
	FlushVisitor(visitorCode string, isUniqueIdentifier ...bool) error

	FlushVisitorInstantly(visitorCode string) error

	FlushAll(instant ...bool)

	// GetFeatureVariationKey returns a variation key for visitor code
	//
	// This method takes a visitorCode and featureKey as mandatory arguments and
	// returns a variation assigned for a given visitor
	// If such a user has never been associated with any feature flag rules, the SDK returns a default variation key
	// You have to make sure that proper error handling is set up in your code as shown in the example to the right to catch potential exceptions.
	//
	// returns FeatureNotFound error
	// returns VisitorCodeNotValid error
	// returns FeatureEnvironmentDisabled error
	//
	// Deprecated: Please use `GetVariation(visitorCode, featureKey, true)` instead
	GetFeatureVariationKey(visitorCode string, featureKey string, isUniqueIdentifier ...bool) (string, error)

	// GetFeatureVariable retrieves a feature variable value from assigned for visitor variation
	//
	// A feature variable can be changed easily via our web application.
	//
	// returns FeatureNotFound error
	// returns VisitorCodeNotValid error
	// returns FeatureEnvironmentDisabled error
	// returns FeatureVariableNotFound error
	// returns VariationNotFound error
	//
	// Deprecated: Please use `GetVariation(visitorCode, featureKey, true)` instead
	GetFeatureVariable(
		visitorCode string, featureKey string, variableKey string, isUniqueIdentifier ...bool,
	) (interface{}, error)

	// IsFeatureActive checks if feature is active for a visitor or not
	// (returns true / false instead of variation key)
	// This method takes a visitorCode and featureKey as mandatory arguments to check
	// if the specified feature will be active for a given user.
	// If such a user has never been associated with this feature flag, the SDK returns a boolean value randomly
	// (true if the user should have this feature or false if not).
	// You have to make sure that proper error handling is set up in your code as shown in the example to the right to catch potential exceptions.
	//
	// returns FeatureNotFound error
	// returns VisitorCodeNotValid error
	IsFeatureActive(visitorCode string, featureKey string, isUniqueIdentifier ...bool) (bool, error)

	IsFeatureActiveWithTracking(visitorCode string, featureKey string, track bool) (bool, error)

	GetVariation(visitorCode string, featureKey string, params ...GetVariationOptParams) (types.Variation, error)

	GetVariations(visitorCode string, params ...GetVariationsOptParams) (map[string]types.Variation, error)

	// GetFeatureVariationVariables retrieves all feature variable values for a given variation
	//
	// This method takes a featureKey and variationKey as mandatory arguments and
	// returns a list of variables for a given variation key
	// A feature variable can be changed easily via our web application.
	//
	// returns FeatureNotFound error
	// returns FeatureEnvironmentDisabled error
	// returns VariationNotFound error
	//
	// Deprecated: Please use `GetVariation(visitorCode, featureKey, false)` instead
	GetFeatureVariationVariables(featureKey string, variationKey string) (map[string]interface{}, error)

	// The GetRemoteData method allows you to retrieve data (according to a key passed as
	// argument)stored on a remote Kameleoon server. Usually data will be stored on our remote servers
	// via the use of our Data API. This method, along with the availability of our highly scalable servers
	// for this purpose, provides a convenient way to quickly store massive amounts of data that
	// can be later retrieved for each of your visitors / users.
	//
	// returns Network timeout error
	GetRemoteData(key string, timeout ...time.Duration) ([]byte, error)

	// GetVisitorWarehouseAudience retrieves data associated with a visitor's warehouse audiences and adds
	// it to the visitor. Retrieves all audience data associated with the visitor in your data warehouse using the
	// specified `visitorCode` and `warehouseKey`. The `warehouseKey` is typically your internal user
	// ID. The `customDataIndex` parameter corresponds to the Kameleoon custom data that Kameleoon uses
	// to target your visitors. You can refer to the
	// warehouse targeting documentation (https://help.kameleoon.com/warehouse-audience-targeting/)
	// for additional details. The method passes the result to the returned `CustomData` object,
	// confirming that the data has been added to the visitor and is available for targeting purposes.
	//
	// Parameters:
	// - VisitorCode: The unique identifier of the visitor for whom you want to retrieve and add the
	//   data.
	// - WarehouseKey: The key to identify the warehouse data, typically your internal user ID. The values is optional.
	// - CustomDataIndex: An integer representing the index of the custom data you want to use to
	//   target your BigQuery Audiences.
	// - Timeout: Time to wait for the response
	//
	// Returns:
	// - A `CustomData` instance confirming that the data has been added to the visitor.
	// - An error if the visitor code is empty or longer than 255 characters.
	GetVisitorWarehouseAudience(params VisitorWarehouseAudienceParams) (*types.CustomData, error)

	// GetVisitorWarehouseAudienceWithOptParams retrieves data associated with a visitor's warehouse audiences and adds
	// it to the visitor. Retrieves all audience data associated with the visitor in your data warehouse using the
	// specified `visitorCode` and `warehouseKey`. The `warehouseKey` is typically your internal user
	// ID. The `customDataIndex` parameter corresponds to the Kameleoon custom data that Kameleoon uses
	// to target your visitors. You can refer to the
	// warehouse targeting documentation (https://help.kameleoon.com/warehouse-audience-targeting/)
	// for additional details. The method passes the result to the returned `CustomData` object,
	// confirming that the data has been added to the visitor and is available for targeting purposes.
	//
	// Parameters:
	// - visitorCode: The unique identifier of the visitor for whom you want to retrieve and add the
	//   data.
	// - customDataIndex: An integer representing the index of the custom data you want to use to
	//   target your BigQuery Audiences.
	// - params: An object with optional parameters: WarehouseKey and Timeout
	//
	// Returns:
	// - A `CustomData` instance confirming that the data has been added to the visitor.
	// - An error if the visitor code is empty or longer than 255 characters.
	GetVisitorWarehouseAudienceWithOptParams(
		visitorCode string, customDataIndex int, params ...VisitorWarehouseAudienceOptParams,
	) (*types.CustomData, error)

	GetRemoteVisitorData(visitorCode string, addData bool, timeout ...time.Duration) ([]types.Data, error)

	// Deprecated: Please use `GetRemoteVisitorDataWithFilter`
	GetRemoteVisitorDataWithOptParams(
		visitorCode string, addData bool, filter types.RemoteVisitorDataFilter, params ...RemoteVisitorDataOptParams,
	) ([]types.Data, error)

	GetRemoteVisitorDataWithFilter(
		visitorCode string, addData bool, filter types.RemoteVisitorDataFilter, params ...RemoteVisitorDataOptParams,
	) ([]types.Data, error)

	OnUpdateConfiguration(handler func())

	// GetFeatureList returns a list of all feature flag keys
	//
	// Deprecated: Please use `GetDataFile()` instead
	GetFeatureList() []string

	// GetActiveFeatureListForVisitor returns a list of active feature flag keys for a visitor
	//
	// returns VisitorCodeNotValid error when visitor code is not valid
	//
	// Deprecated: Please use `GetActiveFeatures`
	GetActiveFeatureListForVisitor(visitorCode string) ([]string, error)

	// GetActiveFeatures returns a map that contains the assigned variations of the active features
	// using the keys of the corresponding active features.
	//
	// returns VisitorCodeNotValid error when visitor code is not valid
	//
	// Deprecated: Please use `GetVariations(visitorCode, true, false)` instead
	GetActiveFeatures(visitorCode string) (map[string]types.Variation, error)

	GetEngineTrackingCode(visitorCode string) string

	// Sets or resets a forced variation for a visitor in a specific experiment,
	// so the experiment will be evaluated to the variation for the visitor.
	//
	// In order to reset the forced variation set the `variationKey` parameter to empty string.
	// If the forced variation you want to reset does not exist, the method will have no effect.
	//
	// Parameters:
	// - visitorCode: The unique visitor code identifying the visitor.
	// - experimentId: The identifier of the experiment you want to set/reset the forced variation for.
	// - variationKey: The identifier of the variation you want the experiment to be evaluated to.
	//   Set to empty string to reset the forced variation.
	// - ForceTargeting: If `true`, the visitor will be targeted to the experiment regardless its conditions.
	//   Otherwise, the normal targeting logic will be preserved. Optional (defaults to `true`).
	//
	// May return one of the following errors:
	// - VisitorCodeInvalid:
	//   The provided visitor code is invalid.
	// - FeatureExperimentNotFound:
	//   The experiment id does not exist in the feature flag.
	// - VariationNotFound:
	//   The variation key does not belong to the experiment.
	SetForcedVariation(
		visitorCode string, experimentId int, variationKey string, params ...SetForcedVariationOptParams,
	) error

	// Evaluates the visitor against all available Audiences Explorer segments and tracks those that match.
	// A detailed analysis of segment performance can then be performed directly in Audiences Explorer.
	//
	// Parameters:
	// - visitorCode: The unique visitor code identifying the visitor.
	//
	// May return one of the following errors:
	// - VisitorCodeInvalid:
	//   The provided visitor code is invalid.
	EvaluateAudiences(visitorCode string) error
}

type kameleoonClient struct {
	cfg                  *KameleoonClientConfig
	visitorManager       storage.VisitorManager
	networkManager       network.NetworkManager
	cookieManager        cookie.CookieManager
	hybridManager        hybrid.HybridManager
	warehouseManager     warehouse.WarehouseManager
	targetingManager     targeting.TargetingManager
	remoteDataManager    remotedata.RemoteDataManager
	trackingManager      tracking.TrackingManager
	configurationManager configuration.ConfigurationManager

	m           sync.Mutex
	readiness   *kameleoonClientReadiness
	dataManager data.DataManager
	closed      bool
}

func newClient(siteCode string, cfg *KameleoonClientConfig) (*kameleoonClient, error) {
	logging.Info("CALL: newClient(siteCode: %s, config: %s)", siteCode, cfg)
	if len(siteCode) == 0 {
		err := errs.NewSiteCodeIsEmpty("Provided siteCode is empty")
		logging.Info("RETURN: newClient(siteCode: %s, config: %s) -> (client, error: %s)",
			siteCode, cfg, err)
		return nil, err
	}
	if err := cfg.defaults(); err != nil {
		logging.Info("RETURN: newClient(siteCode: %s, config: %s) -> (client, error: %s)",
			siteCode, cfg, err)
		return nil, err
	}

	if cfg.Logger != nil {
		logging.SetOldLogger(cfg.Logger)
	}

	if cfg.VerboseMode && logging.GetLogLevel() == logging.WARNING {
		logging.SetLogLevel(logging.INFO)
	}

	df := configuration.NewDataFile(configuration.Configuration{}, "", cfg.Environment)
	dm := data.NewDataManagerImpl(df)
	np := network.NewNetProviderImpl(cfg.Network.ReadTimeout, cfg.Network.WriteTimeout,
		cfg.Network.MaxConnsPerHost, cfg.Network.ProxyURL)
	up := network.NewUrlProviderImpl(siteCode, cfg.NetworkDomain, utils.SdkName, utils.SdkVersion)
	atsf := &network.AccessTokenSourceFactoryImpl{ClientId: cfg.ClientID, ClientSecret: cfg.ClientSecret}
	nm := network.NewNetworkManagerImpl(cfg.Environment, cfg.DefaultTimeout, np, up, atsf)
	vm := newVisitorManager(dm, cfg)
	hm, _ := hybrid.NewHybridManagerImpl(5*time.Second, dm)
	tarM := targeting.NewTargetingManager(dm, vm)
	rdm := remotedata.NewRemoteDataManager(dm, nm, vm)
	trM := tracking.NewTrackingManagerImpl(dm, nm, vm, cfg.TrackingInterval)
	cm := configuration.NewConfigurationManager(dm, nm, &realtime.NetSseClient{}, cfg.RefreshInterval, cfg.Environment)
	client := newClientInternal(cfg, dm, nm, vm, hm, tarM, rdm, trM, cm)
	logging.Info("RETURN: newClient(siteCode: %s, config: %s) -> (client, error: <nil>)",
		siteCode, cfg)
	return client, nil
}

func newClientInternal(
	cfg *KameleoonClientConfig,
	dataManager data.DataManager,
	networkManager network.NetworkManager,
	visitorManager storage.VisitorManager,
	hybridManager hybrid.HybridManager,
	targetingManager targeting.TargetingManager,
	remoteDataManager remotedata.RemoteDataManager,
	trackingManager tracking.TrackingManager,
	configurationManager configuration.ConfigurationManager,
) *kameleoonClient {
	client := &kameleoonClient{
		cfg:                  cfg,
		readiness:            newKameleoonClientReadiness(),
		dataManager:          dataManager,
		visitorManager:       visitorManager,
		hybridManager:        hybridManager,
		networkManager:       networkManager,
		cookieManager:        cookie.NewCookieManagerImpl(dataManager, visitorManager, cfg.TopLevelDomain),
		warehouseManager:     warehouse.NewWarehouseManagerImpl(networkManager, visitorManager),
		targetingManager:     targetingManager,
		remoteDataManager:    remoteDataManager,
		trackingManager:      trackingManager,
		configurationManager: configurationManager,
	}
	go client.updateConfigInitially()
	return client
}

func newVisitorManager(dm data.DataManager, cfg *KameleoonClientConfig) storage.VisitorManager {
	return storage.NewVisitorManagerImpl(dm, cfg.SessionDuration)
}

func (c *kameleoonClient) WaitInit() error {
	logging.Info("CALL: kameleoonClient.WaitInit()")
	err := c.readiness.Wait()
	logging.Info("RETURN: kameleoonClient.WaitInit() -> (error: %s)", err)
	return err
}

func (c *kameleoonClient) close() {
	logging.Debug("CALL: kameleoonClient.close()")
	if !c.closed {
		c.m.Lock()
		if c.closed {
			c.m.Unlock()
		} else {
			c.closed = true
			c.m.Unlock()
			c.visitorManager.Close()
			c.trackingManager.Close()
		}
	}
	logging.Debug("RETURN: kameleoonClient.close()")
}

func (c *kameleoonClient) GetVisitorCode(request *fasthttp.Request, response *fasthttp.Response,
	defaultVisitorCode ...string) (string, error) {
	logging.Info("CALL: kameleoonClient.GetVisitorCode(request, response, defaultVisitorCode: %s)",
		defaultVisitorCode)
	visitorCode, err := c.cookieManager.GetOrAdd(request, response, defaultVisitorCode...)
	logging.Info(
		"RETURN: kameleoonClient.GetVisitorCode(request, response, defaultVisitorCode: %s) -> "+
			"(visitorCode: %s, error: %s)", defaultVisitorCode, visitorCode, err)
	return visitorCode, err
}

func (c *kameleoonClient) SetLegalConsent(visitorCode string, consent bool, response ...*fasthttp.Response) error {
	logging.Info("CALL: kameleoonClient.SetLegalConsent(visitorCode: %s, consent: %s, response)",
		visitorCode, consent)
	err := utils.ValidateVisitorCode(visitorCode)
	if err == nil {
		v := c.visitorManager.GetOrCreateVisitor(visitorCode)
		if consent {
			v.SetLegalConsent(types.LegalConsentGiven)
		} else {
			v.SetLegalConsent(types.LegalConsentNotGiven)
		}
		if len(response) > 0 {
			c.cookieManager.Update(visitorCode, consent, response[0])
		}
	}
	logging.Info("RETURN: kameleoonClient.SetLegalConsent(visitorCode: %s, consent: %s, response) -> (error: %s)",
		visitorCode, consent, err)
	return err
}

func (c *kameleoonClient) AddData(visitorCode string, allData ...types.Data) error {
	return c.AddDataWithOptParams(visitorCode, NewAddDataOptParams(), allData...)
}

func (c *kameleoonClient) AddDataWithOptParams(visitorCode string, optParams AddDataOptParams, allData ...types.Data) error {
	logging.Info("CALL: kameleoonClient.AddDataWithOptParams(visitorCode: %s, track: %t, allData: %s)",
		visitorCode, optParams.track, allData)
	err := utils.ValidateVisitorCode(visitorCode)
	if err == nil {
		c.visitorManager.AddDataWithTrack(visitorCode, optParams.track, allData...)
	}
	logging.Info("RETURN: kameleoonClient.AddDataWithOptParams(visitorCode: %s, track: %t, allData: %s) -> (error: %s)",
		visitorCode, optParams.track, allData, err)
	return err
}

func (c *kameleoonClient) TrackConversion(visitorCode string, goalID int, isUniqueIdentifier ...bool) error {
	logging.Info(
		"CALL: kameleoonClient.TrackConversion(visitorCode: %s, goalID: %s, isUniqueIdentifier: %s)",
		visitorCode, goalID, isUniqueIdentifier)
	var err error
	if len(isUniqueIdentifier) > 0 {
		if err = utils.ValidateVisitorCode(visitorCode); err != nil {
			return err
		}
		c.setUniqueIdentifier(visitorCode, isUniqueIdentifier[0])
	}
	err = c.TrackConversionWithOptParams(visitorCode, goalID, TrackConversionOptParams{})
	logging.Info(
		"RETURN: kameleoonClient.TrackConversion(visitorCode: %s, goalID: %s, isUniqueIdentifier: %s) -> (error: %s)",
		visitorCode, goalID, isUniqueIdentifier, err)
	return err
}

func (c *kameleoonClient) TrackConversionRevenue(
	visitorCode string, goalID int, revenue float64, isUniqueIdentifier ...bool,
) error {
	logging.Info(
		"CALL: kameleoonClient.TrackConversionRevenue(visitorCode: %s, goalID: %s, revenue: %s,"+
			" isUniqueIdentifier: %s)", visitorCode, goalID, revenue, isUniqueIdentifier)
	var err error
	if len(isUniqueIdentifier) > 0 {
		if err = utils.ValidateVisitorCode(visitorCode); err != nil {
			return err
		}
		c.setUniqueIdentifier(visitorCode, isUniqueIdentifier[0])
	}
	err = c.TrackConversionWithOptParams(visitorCode, goalID, TrackConversionOptParams{Revenue: revenue})
	logging.Info(
		"RETURN: kameleoonClient.TrackConversionRevenue(visitorCode: %s, goalID: %s, revenue: %s,"+
			" isUniqueIdentifier: %s) -> (error: %s)", visitorCode, goalID, revenue, isUniqueIdentifier, err)
	return err
}

func (c *kameleoonClient) TrackConversionWithOptParams(
	visitorCode string, goalId int, params TrackConversionOptParams,
) (err error) {
	logging.Info(
		"CALL: kameleoonClient.TrackConversionWithOptParams(visitorCode: %s, goalId: %s, params: %s)",
		visitorCode, goalId, params,
	)
	defer logging.Info(
		"RETURN: kameleoonClient.TrackConversionWithOptParams(visitorCode: %s, goalId: %s, params: %s) -> (error: %s)",
		visitorCode, goalId, params, err,
	)
	if err = utils.ValidateVisitorCode(visitorCode); err != nil {
		return
	}
	paramsConversion := types.ConversionOptParams{
		Revenue:  params.Revenue,
		Negative: params.Negative,
		Metadata: params.Metadata,
	}
	if err = c.AddData(visitorCode, types.NewConversionWithOptParams(goalId, paramsConversion)); err != nil {
		return
	}
	c.trackingManager.AddVisitorCode(visitorCode)
	return
}

func (c *kameleoonClient) FlushVisitor(visitorCode string, isUniqueIdentifier ...bool) error {
	logging.Info("CALL: kameleoonClient.FlushVisitor(visitorCode: %s, isUniqueIdentifier: %s)",
		visitorCode, isUniqueIdentifier)
	err := utils.ValidateVisitorCode(visitorCode)
	if err == nil {
		if len(isUniqueIdentifier) > 0 {
			c.setUniqueIdentifier(visitorCode, isUniqueIdentifier[0])
		}
		c.trackingManager.AddVisitorCode(visitorCode)
	}
	logging.Info("RETURN: kameleoonClient.FlushVisitor(visitorCode: %s, isUniqueIdentifier: %s) -> (error: %s)",
		visitorCode, isUniqueIdentifier, err)
	return err
}

func (c *kameleoonClient) FlushVisitorInstantly(visitorCode string) error {
	logging.Info("CALL: kameleoonClient.FlushVisitorInstantly(visitorCode: %s)", visitorCode)
	err := utils.ValidateVisitorCode(visitorCode)
	if err == nil {
		c.trackingManager.TrackVisitor(visitorCode)
	}
	logging.Info("RETURN: kameleoonClient.FlushVisitorInstantly(visitorCode: %s) -> (error: %s)", visitorCode, err)
	return err
}

func (c *kameleoonClient) FlushAll(instant ...bool) {
	logging.Info("CALL: kameleoonClient.FlushAll(instant: %s)", instant)
	c.visitorManager.Enumerate(func(vc string, v storage.Visitor) bool {
		notEmpty := false
		v.EnumerateSendableData(func(s types.Sendable) bool {
			notEmpty = true
			return false
		})
		if notEmpty {
			c.trackingManager.AddVisitorCode(vc)
		}
		return true
	})
	if (len(instant) > 0) && instant[0] {
		c.trackingManager.TrackAll()
	}
	logging.Info("RETURN: kameleoonClient.FlushAll(instant: %s)", instant)
}

func (c *kameleoonClient) GetFeatureVariationKey(
	visitorCode string, featureKey string, isUniqueIdentifier ...bool,
) (string, error) {
	logging.Info(
		"CALL: kameleoonClient.GetFeatureVariationKey(visitorCode: %s, featureKey: %s, isUniqueIdentifier: %s)",
		visitorCode, featureKey, isUniqueIdentifier)
	if len(isUniqueIdentifier) > 0 {
		c.setUniqueIdentifier(visitorCode, isUniqueIdentifier[0])
	}
	_, variationKey, err := c.getFeatureVariationKey(visitorCode, featureKey)
	logging.Info(
		"RETURN: kameleoonClient.GetFeatureVariationKey(visitorCode: %s, featureKey: %s, isUniqueIdentifier: %s) "+
			"-> (variationKey: %s, error: %s)", visitorCode, featureKey, isUniqueIdentifier, variationKey, err)
	return variationKey, err
}

// getFeatureVariationKey is a helper method for getting variation key for feature flag
func (c *kameleoonClient) getFeatureVariationKey(
	visitorCode string, featureKey string,
) (featureFlag types.IFeatureFlag, variationKey string, err error) {
	logging.Debug(
		"CALL: kameleoonClient.getFeatureVariationKey(visitorCode: %s, featureKey: %s)", visitorCode, featureKey,
	)
	defer func() {
		logging.Debug(
			"RETURN: kameleoonClient.getFeatureVariationKey(visitorCode: %s, featureKey: %s) -> (featureFlag: %s, "+
				"variationKey: %s, error: %s)", visitorCode, featureKey, featureFlag, variationKey, err,
		)
	}()
	if err = utils.ValidateVisitorCode(visitorCode); err != nil {
		variationKey = string(types.VariationOff)
		return
	}
	// find enabled feature flag else return an error
	if featureFlag, err = c.dataManager.DataFile().GetFeatureFlag(featureKey); err != nil {
		variationKey = string(types.VariationOff)
		return
	}
	visitor := c.visitorManager.GetVisitor(visitorCode)
	var evalExp *evaluatedExperiment
	if evalExp, err = c.evaluate(visitor, visitorCode, featureFlag, true, true); err != nil {
		return
	}
	// get variation key from feature flag
	defaultVariationKey := featureFlag.GetDefaultVariationKey()
	variationKey = c.calculateVariationKey(evalExp, defaultVariationKey)
	c.trackingManager.AddVisitorCode(visitorCode)
	return
}

func (c *kameleoonClient) saveVariation(
	visitorCode string, evalExp *evaluatedExperiment, track bool,
) {
	if (evalExp == nil) || (evalExp.experiment.ExperimentId == 0) || (evalExp.varByExp.VariationID == nil) {
		return
	}
	logging.Debug(
		"CALL: kameleoonClient.saveVariation(visitorCode: %s, evalExp: %s, track: %s)",
		visitorCode, evalExp, track,
	)
	visitor := c.visitorManager.GetOrCreateVisitor(visitorCode)
	asVariation := types.NewAssignedVariation(
		evalExp.experiment.ExperimentId, *evalExp.varByExp.VariationID, evalExp.ruleType,
	)
	if !track {
		asVariation.MarkAsSent()
	}
	visitor.AssignVariation(asVariation)
	logging.Debug(
		"RETURN: kameleoonClient.saveVariation(visitorCode: %s, evalExp: %s, track: %s)",
		visitorCode, evalExp, track,
	)
}

func (c *kameleoonClient) calculateVariationKey(evalExp *evaluatedExperiment, defaultVariationKey string) string {
	logging.Debug(
		"CALL: kameleoonClient.calculateVariationKey(evalExp: %s, defaultVariationKey: %s)",
		evalExp, defaultVariationKey,
	)
	var variationKey string
	if evalExp != nil {
		variationKey = evalExp.varByExp.VariationKey
	} else {
		variationKey = defaultVariationKey
	}
	logging.Debug(
		"RETURN: kameleoonClient.calculateVariationKey(evalExp: %s, defaultVariationKey: %s) -> "+
			"(variationKey: %s)", evalExp, defaultVariationKey, variationKey,
	)
	return variationKey
}

func getCodeForHash(visitor storage.Visitor, visitorCode string, bucketingCustomDataIndex *int) string {
	if visitor != nil {
		// 1. Try to use the bucketing custom data's value if bucketingCustomDataIndex is defined
		if bucketingCustomDataIndex != nil {
			bucketingCustomData := visitor.CustomData().Get(*bucketingCustomDataIndex)
			if (bucketingCustomData != nil) && (len(bucketingCustomData.Values()) > 0) {
				return bucketingCustomData.Values()[0]
			}
		}
		// 2. Use mappingIdentifier instead of visitorCode if it was set up
		if visitor.MappingIdentifier() != nil {
			return *visitor.MappingIdentifier()
		}
	}
	return visitorCode
}

func (c *kameleoonClient) evaluateCBScores(
	visitor storage.Visitor, visitorCode string, rule types.IRule, bucketingCustomDataIndex *int,
) *evaluatedExperiment {
	if (visitor == nil) || (visitor.CBScores() == nil) {
		return nil
	}
	logging.Debug(
		"CALL: kameleoonClient.evaluateCBScores(visitor, visitorCode: %s, rule: %s, bucketingCustomDataIndex: %s)",
		visitorCode, rule, bucketingCustomDataIndex,
	)
	var evalExp *evaluatedExperiment
	ruleVarsByExp := rule.GetRuleBase().VariationsByExposition
	if varIdGroupByScores, ok := visitor.CBScores().Values()[rule.GetRuleBase().ExperimentId]; ok {
		var varByExpInCbs []*types.VariationByExposition
		for _, varGroup := range varIdGroupByScores {
			// Finding varByExps which exist in CBS variation IDs
			varByExpInCbs = make([]*types.VariationByExposition, 0, len(ruleVarsByExp))
			for i := 0; i < len(ruleVarsByExp); i++ {
				if (ruleVarsByExp[i].VariationID != nil) &&
					utils.SliceContains(varGroup.Ids(), *ruleVarsByExp[i].VariationID) {
					varByExpInCbs = append(varByExpInCbs, &ruleVarsByExp[i])
				}
			}
			if len(varByExpInCbs) > 0 { // Skiping if not found any varByExp
				break // We need take only one list with the highest scores
			}
		}
		if len(varByExpInCbs) > 0 {
			var idx int
			if len(varByExpInCbs) > 1 { // if more than one varByExp for score -> randomly get
				codeForHash := getCodeForHash(visitor, visitorCode, bucketingCustomDataIndex)
				variationHash := utils.ObtainHashRule(
					codeForHash, rule.GetRuleBase().ExperimentId, rule.GetRuleBase().RespoolTime,
				)
				logging.Debug("Calculated CBS hash %s for code %s", variationHash, codeForHash)
				idx = int(variationHash * float64(len(varByExpInCbs)))
				if idx >= len(varByExpInCbs) {
					idx = len(varByExpInCbs) - 1
				}
			}
			evalExp = newEvaluatedExperimentFromVarByExpRule(varByExpInCbs[idx], rule)
		}
	}
	logging.Debug(
		"RETURN: kameleoonClient.evaluateCBScores(visitor, visitorCode: %s, rule: %s, bucketingCustomDataIndex: %s)"+
			" -> (evalExp: %s)", visitorCode, rule, bucketingCustomDataIndex, evalExp,
	)
	return evalExp
}

// getVariationRuleForFeature is a helper method for calculate variation key for feature flag
func (c *kameleoonClient) calculateVariationRuleForFeature(
	visitorCode string, featureFlag types.IFeatureFlag,
) (evalExp *evaluatedExperiment, err error) {
	logging.Debug(
		"CALL: kameleoonClient.calculateVariationRuleForFeature(visitorCode: %s, featureFlag: %s)",
		visitorCode, featureFlag,
	)
	defer func() {
		logging.Debug(
			"RETURN: kameleoonClient.calculateVariationRuleForFeature(visitorCode: %s, featureFlag: %s)"+
				" -> (evalExp: %s, err: %s)",
			visitorCode, featureFlag, evalExp, err,
		)
	}()
	visitor := c.visitorManager.GetVisitor(visitorCode)
	consent, blockingBehaviour := c.getConsentAndBlockingBehaviour(visitor)
	codeForHash := getCodeForHash(visitor, visitorCode, featureFlag.GetBucketingCustomDataIndex())
	// no rules -> return DefaultVariationKey
	for _, rule := range featureFlag.GetRules() {
		var forcedVariation *types.ForcedExperimentVariation
		if visitor != nil {
			forcedVariation = visitor.GetForcedExperimentVariation(rule.GetRuleBase().ExperimentId)
			if (forcedVariation != nil) && forcedVariation.ForceTargeting() {
				// Forcing experiment variation in force-targeting mode
				return newEvaluatedExperimentFromVarByExpRule(forcedVariation.VarByExp(), rule), nil
			}
		}
		// check if visitor is targeted for rule, else next rule
		if !c.targetingManager.CheckTargeting(visitorCode, rule.GetRuleBase().ExperimentId, rule.GetTargetingSegment()) {
			continue
		}
		if forcedVariation != nil {
			// Forcing experiment variation in targeting-only mode
			return newEvaluatedExperimentFromVarByExpRule(forcedVariation.VarByExp(), rule), nil
		}

		// Disable searching in variation storage (uncommented if you need use variation storage)
		// check for saved variation for rule if it's experimentation rule
		// if savedVariation, found := c.getSavedVariationForRule(visitorCode, &rule); found {
		// 	return savedVariation, &rule, false
		// }

		// used for rule exposition
		hashRule := utils.ObtainHashRule(codeForHash, rule.GetRuleBase().Id, rule.GetRuleBase().RespoolTime)
		logging.Debug("Calculated rule hash %s for code %s", hashRule, codeForHash)
		// check main expostion for rule with hashRule
		if hashRule <= rule.GetRuleBase().Exposition {
			// Checking if the evaluation is blocked due to the consent policy
			if (consent == types.LegalConsentNotGiven) && (rule.GetRuleBase().Type == types.RuleTypeExperimentation) {
				if blockingBehaviour == types.PartiallyBlockedByConsent {
					return nil, nil
				}
				// TODO: In the next major, create a new error type for the Completely Blocked by Consent error
				return nil, errs.NewFeatureEnvironmentDisabledWithMessage(fmt.Sprintf(
					"Evaluation of %v is blocked because consent is not provided for visitor '%s'",
					rule, visitorCode,
				))
			}
			// check main exposition for rule with hashRule
			evalExp = c.evaluateCBScores(visitor, visitorCode, rule, featureFlag.GetBucketingCustomDataIndex())
			if evalExp != nil {
				return
			}
			if rule.IsTargetDeliveryType() {
				var variation *types.VariationByExposition
				if len(rule.GetRuleBase().VariationsByExposition) > 0 {
					variation = &rule.GetRuleBase().VariationsByExposition[0]
				}
				return newEvaluatedExperimentFromVarByExpRule(variation, rule), nil
			}
			// used for variation's expositions
			hashVariation := utils.ObtainHashRule(
				codeForHash, rule.GetRuleBase().ExperimentId, rule.GetRuleBase().RespoolTime,
			)
			logging.Debug("Calculated variation hash %s for code %s", hashVariation, codeForHash)
			// get variation with new hashVariation
			variation := rule.GetVariationByHash(hashVariation)
			if variation != nil {
				return newEvaluatedExperimentFromVarByExpRule(variation, rule), nil
			}
		}
		if rule.IsTargetDeliveryType() {
			break
		}
	}
	return nil, nil
}

func (c *kameleoonClient) getConsentAndBlockingBehaviour(visitor storage.Visitor) (types.LegalConsent, types.ConsentBlockingBehaviour) {
	dataFile := c.dataManager.DataFile()

	consent := types.LegalConsentGiven
	if dataFile.Settings().IsConsentRequired() {
		consent = types.LegalConsentUnknown
		if visitor != nil {
			consent = visitor.LegalConsent()
		}
	}
	behaviour := dataFile.Settings().BlockingBehaviourIfConsentNotGiven()

	return consent, behaviour
}

// func (c *kameleoonClient) getSavedVariationForRule(visitorCode string, rule *configuration.Rule) (*types.VariationByExposition, bool) {
// 	if (rule != nil) && rule.IsExperimentType() && (rule.ExperimentId != 0) {
// 		v := c.visitorManager.GetVisitor(visitorCode)
// 		if v != nil {
// 			if variation := v.Variations().Get(rule.ExperimentId); variation != nil {
// 				return rule.GetVariation(variation.VariationId()), true
// 			}
// 		}
// 	}
// 	return nil, false
// }

func (c *kameleoonClient) GetFeatureVariable(
	visitorCode string, featureKey string, variableKey string, isUniqueIdentifier ...bool,
) (interface{}, error) {
	logging.Info(
		"CALL: kameleoonClient.GetFeatureVariable(visitorCode: %s, featureKey: %s, variableKey: %s,"+
			" isUniqueIdentifier: %s)", visitorCode, featureKey, variableKey, isUniqueIdentifier)
	if len(isUniqueIdentifier) > 0 {
		c.setUniqueIdentifier(visitorCode, isUniqueIdentifier[0])
	}
	var variableValue interface{}
	featureFlag, variationKey, err := c.getFeatureVariationKey(visitorCode, featureKey)
	if err == nil {
		variation, exist := featureFlag.GetVariationByKey(variationKey)
		if !exist {
			err = errs.NewFeatureVariationNotFound(featureKey, variationKey)
		} else {
			variable, exist := variation.GetVariableByKey(variableKey)
			if !exist {
				err = errs.NewFeatureVariableNotFound(featureKey, variationKey, variableKey)
			} else {
				variableValue = variable.GetVariableValue()
			}
		}
	}

	logging.Info(
		"RETURN: kameleoonClient.GetFeatureVariable(visitorCode: %s, featureKey: %s, variableKey: %s, "+
			"isUniqueIdentifier: %s) -> (variable: %s, err: %s)",
		visitorCode, featureKey, variableKey, isUniqueIdentifier, variableValue, err)
	return variableValue, err
}

func (c *kameleoonClient) IsFeatureActive(
	visitorCode string, featureKey string, isUniqueIdentifier ...bool,
) (isFeatureActive bool, err error) {
	logging.Info(
		"CALL: kameleoonClient.IsFeatureActive(visitorCode: %s, featureKey: %s, isUniqueIdentifier: %s)",
		visitorCode, featureKey, isUniqueIdentifier,
	)
	defer func() {
		logging.Info(
			"RETURN: kameleoonClient.IsFeatureActive(visitorCode: %s, featureKey: %s, isUniqueIdentifier: %s) -> "+
				"(isFeatureActive: %s, err: %s)", visitorCode, featureKey, isUniqueIdentifier, isFeatureActive, err,
		)
	}()
	if err = utils.ValidateVisitorCode(visitorCode); err != nil {
		return
	}
	if len(isUniqueIdentifier) > 0 {
		c.setUniqueIdentifier(visitorCode, isUniqueIdentifier[0])
	}
	isFeatureActive, err = c.isFeatureActive(visitorCode, featureKey, true)
	return
}

func (c *kameleoonClient) IsFeatureActiveWithTracking(
	visitorCode string, featureKey string, track bool,
) (isFeatureActive bool, err error) {
	logging.Info(
		"CALL: kameleoonClient.IsFeatureActiveWithTracking(visitorCode: %s, featureKey: %s, track: %s)",
		visitorCode, featureKey, track,
	)
	defer func() {
		logging.Info(
			"RETURN: kameleoonClient.IsFeatureActiveWithTracking(visitorCode: %s, featureKey: %s, track: %s) -> "+
				"(isFeatureActive: %s, err: %s)", visitorCode, featureKey, track, isFeatureActive, err,
		)
	}()
	if err = utils.ValidateVisitorCode(visitorCode); err != nil {
		return
	}
	isFeatureActive, err = c.isFeatureActive(visitorCode, featureKey, track)
	return
}

func (c *kameleoonClient) isFeatureActive(
	visitorCode string, featureKey string, track bool,
) (isFeatureActive bool, err error) {
	logging.Debug(
		"CALL: kameleoonClient.isFeatureActive(visitorCode: %s, featureKey: %s, track: %s)",
		visitorCode, featureKey, track,
	)
	defer func() {
		logging.Debug(
			"RETURN: kameleoonClient.isFeatureActive(visitorCode: %s, featureKey: %s, track: %s) -> "+
				"(isFeatureActive: %s, err: %s)", visitorCode, featureKey, track, isFeatureActive, err,
		)
	}()
	ignoreFeatureEnvDisabled := func(err error) (bool, error) {
		if err == nil {
			return true, nil
		}
		if _, ok := err.(*errs.FeatureEnvironmentDisabled); ok {
			return false, nil
		}
		return false, err
	}
	var featureFlag types.IFeatureFlag
	featureFlag, err = c.dataManager.DataFile().GetFeatureFlag(featureKey)
	if ok, err := ignoreFeatureEnvDisabled(err); !ok {
		return false, err
	}
	var variationKey string
	variationKey, _, err = c.getVariationInfo(visitorCode, featureFlag, track)
	if ok, err := ignoreFeatureEnvDisabled(err); !ok {
		return false, err
	}
	isFeatureActive = variationKey != string(types.VariationOff)
	if track {
		c.trackingManager.AddVisitorCode(visitorCode)
	}
	return
}

func (c *kameleoonClient) GetVariation(
	visitorCode string, featureKey string, params ...GetVariationOptParams,
) (externalVariation types.Variation, err error) {
	logging.Info(
		"CALL: kameleoonClient.GetVariation(visitorCode: %s, featureKey: %s, params: %s)",
		visitorCode, featureKey, params,
	)
	defer func() {
		logging.Info(
			"RETURN: kameleoonClient.GetVariation(visitorCode: %s, featureKey: %s, params: %s) -> "+
				"(variation: %s, err: %s)", visitorCode, featureKey, params, externalVariation, err,
		)
	}()
	var p GetVariationOptParams
	if len(params) > 0 {
		p = params[0]
	} else {
		p = NewGetVariationOptParams()
	}
	if err = utils.ValidateVisitorCode(visitorCode); err != nil {
		return
	}
	var featureFlag types.IFeatureFlag
	if featureFlag, err = c.dataManager.DataFile().GetFeatureFlag(featureKey); err != nil {
		return
	}
	var variationKey string
	var evalExp *evaluatedExperiment
	if variationKey, evalExp, err = c.getVariationInfo(visitorCode, featureFlag, p.track); err != nil {
		return
	}
	variation, _ := featureFlag.GetVariationByKey(variationKey)
	externalVariation = createExternalVariation(variation, evalExp)
	if p.track {
		c.trackingManager.AddVisitorCode(visitorCode)
	}
	return
}

func (c *kameleoonClient) GetVariations(
	visitorCode string, params ...GetVariationsOptParams,
) (variations map[string]types.Variation, err error) {
	logging.Info(
		"CALL: kameleoonClient.GetVariations(visitorCode: %s, params: %s)",
		visitorCode, params,
	)
	defer func() {
		logging.Info(
			"RETURN: kameleoonClient.GetVariations(visitorCode: %s, params: %s) -> (variations: %s, err: %s)",
			visitorCode, params, variations, err,
		)
	}()
	var p GetVariationsOptParams
	if len(params) > 0 {
		p = params[0]
	} else {
		p = NewGetVariationsOptParams()
	}
	if err = utils.ValidateVisitorCode(visitorCode); err != nil {
		return
	}
	variations = make(map[string]types.Variation)
	for _, ff := range c.dataManager.DataFile().GetOrderedFeatureFlags() {
		var variationKey string
		var evalExp *evaluatedExperiment
		variationKey, evalExp, err = c.getVariationInfo(visitorCode, ff, p.track)
		if err == nil {
			if p.onlyActive && (variationKey == string(types.VariationOff)) {
				continue
			}
			variation, _ := ff.GetVariationByKey(variationKey)
			variations[ff.GetFeatureKey()] = createExternalVariation(variation, evalExp)
		} else {
			switch err.(type) {
			case *errs.FeatureEnvironmentDisabled:
				continue
			}
			return nil, err
		}
	}
	if p.track {
		c.trackingManager.AddVisitorCode(visitorCode)
	}
	return
}

func (c *kameleoonClient) getVariationInfo(
	visitorCode string, featureFlag types.IFeatureFlag, track bool,
) (variationKey string, evalExp *evaluatedExperiment, err error) {
	logging.Debug(
		"CALL: kameleoonClient.getVariationInfo(visitorCode: %s, featureFlag: %s, track: %s)",
		visitorCode, featureFlag, track,
	)
	visitor := c.visitorManager.GetVisitor(visitorCode)
	evalExp, err = c.evaluate(visitor, visitorCode, featureFlag, track, true)
	if err == nil {
		defaultVariationKey := featureFlag.GetDefaultVariationKey()
		variationKey = c.calculateVariationKey(evalExp, defaultVariationKey)
	}
	logging.Debug(
		"RETURN: kameleoonClient.getVariationInfo(visitorCode: %s, featureFlag: %s, track: %s)"+
			" -> (variationKey: %s, evalExp: %s, err: %s)", visitorCode, featureFlag, track, variationKey, evalExp, err,
	)
	return
}

func (c *kameleoonClient) evaluate(
	visitor storage.Visitor, visitorCode string, featureFlag types.IFeatureFlag, track, save bool,
) (evalExp *evaluatedExperiment, err error) {
	logging.Debug(
		"CALL: kameleoonClient.evaluate(visitor, visitorCode: %s, featureFlag: %s, track: %s, save: %s)",
		visitorCode, featureFlag, track, save,
	)
	defer func() {
		logging.Debug(
			"RETURN: kameleoonClient.evaluate(visitor, visitorCode: %s, featureFlag: %s, track: %s, save: %s)"+
				" -> (evalExp: %s, err: %s)", visitorCode, featureFlag, track, save, evalExp, err,
		)
	}()
	var forcedVariation *types.ForcedFeatureVariation
	if visitor != nil {
		forcedVariation = visitor.GetForcedFeatureVariation(featureFlag.GetFeatureKey())
	}

	if forcedVariation == nil || !forcedVariation.Simulated() {
		if err = c.dataManager.DataFile().EnsureEnvironmentEnabled(featureFlag); err != nil {
			return
		}
	}

	if forcedVariation != nil {
		evalExp = newEvaluatedExperimentFromForcedVariation(forcedVariation)
	} else {
		var isVisitorNotInHoldout bool
		isVisitorNotInHoldout, err = c.isVisitorNotInHoldout(
			visitor, visitorCode, track, save, featureFlag.GetBucketingCustomDataIndex(),
		)
		if err != nil {
			return
		}
		if isVisitorNotInHoldout && c.isFFUnrestrictedByMEGroup(visitor, visitorCode, featureFlag) {
			if evalExp, err = c.calculateVariationRuleForFeature(visitorCode, featureFlag); err != nil {
				return
			}
		}
	}
	if save && ((forcedVariation == nil) || !forcedVariation.Simulated()) {
		c.saveVariation(visitorCode, evalExp, track)
	}
	return
}

func (c *kameleoonClient) isFFUnrestrictedByMEGroup(
	visitor storage.Visitor, visitorCode string, featureFlag types.IFeatureFlag,
) bool {
	meGroupName := featureFlag.GetMEGroupName()
	if meGroupName == "" {
		return true
	}
	logging.Debug(
		"CALL: kameleoonClient.isFFUnrestrictedByMEGroup(visitor, visitorCode: %s, featureFlag: %s)",
		visitorCode, featureFlag,
	)
	unrestricted := true
	if meGroup := c.dataManager.DataFile().MEGroups()[meGroupName]; meGroup != nil {
		codeForHash := getCodeForHash(visitor, visitorCode, featureFlag.GetBucketingCustomDataIndex())
		meGroupHash := utils.ObtainHashForMEGroup(codeForHash, meGroupName)
		logging.Debug("Calculated ME group hash %s for code: %s, meGroup: %s", meGroupHash, codeForHash, meGroupName)
		unrestricted = meGroup.GetFeatureFlagByHash(meGroupHash) == featureFlag
	}
	logging.Debug(
		"RETURN: kameleoonClient.isFFUnrestrictedByMEGroup(visitor, visitorCode: %s, featureFlag: %s)"+
			" -> (unrestricted: %s)", visitorCode, featureFlag, unrestricted,
	)
	return unrestricted
}

func (c *kameleoonClient) isVisitorNotInHoldout(
	visitor storage.Visitor, visitorCode string, track, save bool, bucketingCustomDataIndex *int,
) (isNotInHoldout bool, err error) {
	holdout := c.dataManager.DataFile().Holdout()
	isNotInHoldout = true
	if holdout == nil {
		return
	}
	logging.Debug(
		"CALL: kameleoonClient.isVisitorNotInHoldout(visitor, visitorCode: %s, track: %s, save: %s,"+
			" bucketingCustomDataIndex: %s)", visitorCode, track, save, bucketingCustomDataIndex,
	)
	consent, blockingBehaviour := c.getConsentAndBlockingBehaviour(visitor)
	if consent == types.LegalConsentNotGiven && blockingBehaviour == types.CompletelyBlockedByConsent {
		err = errs.NewFeatureEnvironmentDisabledWithMessage("Evaluation for a holdout is blocked because visitor's consent was not provided.")
		return
	}
	const inHoldoutVariationKey = "in-holdout"
	codeForHash := getCodeForHash(visitor, visitorCode, bucketingCustomDataIndex)
	variationHash := utils.ObtainHash(codeForHash, holdout.ExperimentId)
	logging.Debug("Calculated holdout hash %s for code %s", variationHash, codeForHash)
	if varByExp := holdout.GetVariationByHash(variationHash); varByExp != nil {
		isNotInHoldout = varByExp.VariationKey != inHoldoutVariationKey
		if save {
			evalExp := &evaluatedExperiment{
				varByExp:   varByExp,
				experiment: holdout,
				ruleType:   types.RuleTypeExperimentation,
			}
			c.saveVariation(visitorCode, evalExp, track)
		}
	}
	logging.Debug(
		"RETURN: kameleoonClient.isVisitorNotInHoldout(visitor, visitorCode: %s, track: %s, save: %s,"+
			" bucketingCustomDataIndex: %s) -> (isNotInHoldout: %s)",
		visitorCode, track, save, bucketingCustomDataIndex, isNotInHoldout,
	)
	return
}

func createExternalVariation(
	internalVariation *types.VariationFeatureFlag, evalExp *evaluatedExperiment,
) (variation types.Variation) {
	logging.Debug("CALL: createExternalVariation(internalVariation: %s, evalExp: %s)", internalVariation, evalExp)
	defer func() {
		logging.Debug(
			"RETURN: createExternalVariation(internalVariation: %s, evalExp: %s) -> (variation: %s)",
			internalVariation, evalExp, variation,
		)
	}()
	var variationId *int
	var experimentId *int
	if evalExp != nil {
		variationId = utils.Reref(evalExp.varByExp.VariationID)
		experimentId = utils.Reref(&evalExp.experiment.ExperimentId)
	}
	variation = (types.Variation{}).BuildFromInternal(internalVariation, variationId, experimentId)
	return
}

func (c *kameleoonClient) GetFeatureVariationVariables(
	featureKey string, variationKey string,
) (map[string]interface{}, error) {
	logging.Info(
		"CALL: kameleoonClient.GetFeatureVariationVariables(featureKey: %s, variationKey: %s)",
		featureKey, variationKey)
	var mapVariableValues map[string]interface{}
	featureFlag, err := c.dataManager.DataFile().GetFeatureFlag(featureKey)
	if err == nil {
		if err = c.dataManager.DataFile().EnsureEnvironmentEnabled(featureFlag); err != nil {
			return nil, err
		}
		variation, exist := featureFlag.GetVariationByKey(variationKey)
		if !exist {
			err = errs.NewFeatureVariationNotFound(featureKey, variationKey)
		} else {
			mapVariableValues = make(map[string]interface{})
			for _, variable := range variation.Variables {
				mapVariableValues[variable.Key] = variable.GetVariableValue()
			}
		}
	}
	logging.Info(
		"RETURN: kameleoonClient.GetFeatureVariationVariables(featureKey: %s, variationKey: %s) -> "+
			"(variables: %s, error: %s)", featureKey, variationKey, mapVariableValues, err)
	return mapVariableValues, err
}

func (c *kameleoonClient) GetRemoteData(key string, timeout ...time.Duration) ([]byte, error) {
	logging.Info("CALL: kameleoonClient.GetRemoteData(key: %s, timeout: %s)", key, timeout)
	remoteData, err := c.remoteDataManager.GetData(key, timeout...)
	logging.Info("RETURN: kameleoonClient.GetRemoteData(key: %s, timeout: %s) -> (remoteData: %s, error: %s)",
		remoteData, err)
	return remoteData, err
}

func (c *kameleoonClient) GetRemoteVisitorData(
	visitorCode string,
	addData bool,
	timeout ...time.Duration,
) ([]types.Data, error) {
	logging.Info("CALL: kameleoonClient.GetRemoteVisitorData(visitorCode: %s, addData: %s, timeout: %s)",
		visitorCode, addData, timeout)
	filter := types.DefaultRemoteVisitorDataFilter()
	visitorData, err := c.remoteDataManager.GetVisitorData(visitorCode, filter, addData, timeout...)
	logging.Info(
		"RETURN: kameleoonClient.GetRemoteVisitorData(visitorCode: %s, addData: %s, timeout: %s) -> "+
			"(visitorData: %s, error: %s)", visitorCode, addData, timeout, visitorData, err)
	return visitorData, err
}

func (c *kameleoonClient) GetRemoteVisitorDataWithOptParams(
	visitorCode string, addData bool, filter types.RemoteVisitorDataFilter, params ...RemoteVisitorDataOptParams,
) ([]types.Data, error) {
	logging.Info(
		"CALL: kameleoonClient.GetRemoteVisitorDataWithOptParams(visitorCode: %s, addData: %s, filter: %s, params: %s)",
		visitorCode, addData, filter, params)
	var p RemoteVisitorDataOptParams
	if len(params) > 0 {
		p = params[0]
	}
	var timeout []time.Duration
	if p.Timeout > 0 {
		timeout = []time.Duration{p.Timeout}
	}
	c.setUniqueIdentifier(visitorCode, p.IsUniqueIdentifier)
	visitorData, err := c.remoteDataManager.GetVisitorData(visitorCode, filter, addData, timeout...)
	logging.Info(
		"RETURN: kameleoonClient.GetRemoteVisitorDataWithOptParams(visitorCode: %s, addData: %s, filter: %s, "+
			"params: %s) -> (visitorData: %s, error: %s)", visitorCode, addData, filter, params, visitorData, err)
	return visitorData, err
}

func (c *kameleoonClient) GetRemoteVisitorDataWithFilter(
	visitorCode string, addData bool, filter types.RemoteVisitorDataFilter, params ...RemoteVisitorDataOptParams,
) ([]types.Data, error) {
	logging.Info(
		"CALL: kameleoonClient.GetRemoteVisitorDataWithFilter(visitorCode: %s, addData: %s, filter: %s, params: %s)",
		visitorCode, addData, filter, params)
	var p RemoteVisitorDataOptParams
	if len(params) > 0 {
		p = params[0]
	}
	var timeout []time.Duration
	if p.Timeout > 0 {
		timeout = []time.Duration{p.Timeout}
	}
	remoteVisitorData, err := c.remoteDataManager.GetVisitorData(visitorCode, filter, addData, timeout...)
	logging.Info(
		"RETURN: kameleoonClient.GetRemoteVisitorDataWithFilter(visitorCode: %s, addData: %s, filter: %s,"+
			" params: %s) -> (remoteVisitorData: %s, error: %s)",
		visitorCode, addData, filter, params, remoteVisitorData, err)
	return remoteVisitorData, err
}

func (c *kameleoonClient) updateConfigInitially() {
	logging.Debug("CALL: kameleoonClient.updateConfigInitially()")
	err := c.configurationManager.Start()
	c.readiness.set(err)
	logging.Debug("RETURN: kameleoonClient.updateConfigInitially()")
}

func (c *kameleoonClient) OnUpdateConfiguration(handler func()) {
	c.configurationManager.OnUpdateConfiguration(handler)
	logging.Info("CALL/RETURN: kameleoonClient.OnUpdateConfiguration(handler)")
}

/*
func (c *kameleoonClient) getValidSavedVariation(visitorCode string, experiment *configuration.Experiment) (int, bool) {
	//get saved variation
	if savedVariationId, exist := c.variationStorage.GetVariationId(visitorCode, experiment.ID); exist {
		// get actual respoolTime value for saved variation
		respoolTimeValue := 0
		for _, respoolTime := range experiment.RespoolTime {
			if respoolTime.VariationId == savedVariationId {
				respoolTimeValue = int(respoolTime.Value)
			}
		}
		// checking variation for validity along with actual respoolTime
		return c.variationStorage.IsVariationValid(visitorCode, experiment.ID, respoolTimeValue)
	}
	return 0, false
}
//*/

// Deprecated: Please use `GetDataFile()` instead.
func (c *kameleoonClient) GetFeatureList() []string {
	logging.Info("CALL: kameleoonClient.GetFeatureList()")
	featureFlags := c.dataManager.DataFile().GetFeatureFlags()
	arrayKeys := make([]string, 0, len(featureFlags))
	for _, ff := range featureFlags {
		arrayKeys = append(arrayKeys, ff.GetFeatureKey())
	}
	logging.Info("RETURN: kameleoonClient.GetFeatureList() -> (features: %s)", arrayKeys)
	return arrayKeys
}

func (c *kameleoonClient) GetActiveFeatureListForVisitor(visitorCode string) (arrayIds []string, err error) {
	logging.Info("CALL: kameleoonClient.GetActiveFeatureListForVisitor(visitorCode: %s)", visitorCode)
	defer func() {
		logging.Info(
			"RETURN: kameleoonClient.GetActiveFeatureListForVisitor(visitorCode: %s) -> (activeFeatures: %s, err: %s)",
			visitorCode, arrayIds, err,
		)
	}()
	err = utils.ValidateVisitorCode(visitorCode)
	if err == nil {
		visitor := c.visitorManager.GetVisitor(visitorCode)
		featureFlags := c.dataManager.DataFile().GetOrderedFeatureFlags()
		arrayIds = make([]string, 0, len(featureFlags))
		for _, ff := range featureFlags {
			var evalExp *evaluatedExperiment
			evalExp, err = c.evaluate(visitor, visitorCode, ff, false, false)
			if err == nil {
				variationKey := c.calculateVariationKey(evalExp, ff.GetDefaultVariationKey())
				if variationKey != string(types.VariationOff) {
					arrayIds = append(arrayIds, ff.GetFeatureKey())
				}
			} else {
				switch err.(type) {
				case *errs.FeatureEnvironmentDisabled:
					continue
				}
				return nil, err
			}
		}
	} else {
		arrayIds = []string{}
	}
	return arrayIds, err
}

func (c *kameleoonClient) GetActiveFeatures(visitorCode string) (activeFeatures map[string]types.Variation, err error) {
	logging.Info("CALL: kameleoonClient.GetActiveFeatures(visitorCode: %s)", visitorCode)
	defer func() {
		logging.Info(
			"RETURN: kameleoonClient.GetActiveFeatures(visitorCode: %s) -> (activeFeatures: %s, error: %s)",
			visitorCode, activeFeatures, err,
		)
	}()
	if err = utils.ValidateVisitorCode(visitorCode); err != nil {
		return
	}
	visitor := c.visitorManager.GetVisitor(visitorCode)
	activeFeatures = make(map[string]types.Variation)
	for _, ff := range c.dataManager.DataFile().GetOrderedFeatureFlags() {
		var evalExp *evaluatedExperiment
		evalExp, err = c.evaluate(visitor, visitorCode, ff, false, false)
		if err == nil {
			variationKey := c.calculateVariationKey(evalExp, ff.GetDefaultVariationKey())
			if variationKey == string(types.VariationOff) {
				continue
			}
			variation, _ := ff.GetVariationByKey(variationKey)
			activeFeatures[ff.GetFeatureKey()] = createExternalVariation(variation, evalExp)
		} else {
			switch err.(type) {
			case *errs.FeatureEnvironmentDisabled:
				continue
			}
			return nil, err
		}
	}
	return
}

func (c *kameleoonClient) GetEngineTrackingCode(visitorCode string) string {
	logging.Info("CALL: kameleoonClient.GetEngineTrackingCode(visitorCode: %s)", visitorCode)
	var engineTrackingCode string
	if c.hybridManager == nil {
		logging.Error("HybridManager wasn't initialized properly. GetEngineTrackingCode method isn't avaiable")
		engineTrackingCode = ""
	} else {
		visitor := c.visitorManager.GetVisitor(visitorCode)
		var variations storage.DataMapStorage[int, *types.AssignedVariation]
		if visitor != nil {
			variations = visitor.Variations()
		}
		engineTrackingCode = c.hybridManager.GetEngineTrackingCode(variations)
	}
	logging.Info("RETURN: kameleoonClient.GetEngineTrackingCode(visitorCode: %s) -> (engineTrackingCode: %s)",
		visitorCode, engineTrackingCode)
	return engineTrackingCode
}

func (c *kameleoonClient) GetVisitorWarehouseAudience(params VisitorWarehouseAudienceParams) (*types.CustomData, error) {
	logging.Info("CALL: kameleoonClient.GetVisitorWarehouseAudience(params: %s)", params)
	customData, err := c.warehouseManager.GetVisitorWarehouseAudience(
		params.VisitorCode, params.WarehouseKey, params.CustomDataIndex, params.Timeout)
	logging.Info("RETURN: kameleoonClient.GetVisitorWarehouseAudience(params: %s) -> (customData: %s, error: %s)",
		params, customData, err)
	return customData, err
}

func (c *kameleoonClient) GetVisitorWarehouseAudienceWithOptParams(
	visitorCode string, customDataIndex int, params ...VisitorWarehouseAudienceOptParams,
) (*types.CustomData, error) {
	logging.Info(
		"CALL: kameleoonClient.GetVisitorWarehouseAudienceWithOptParams(visitorCode: %s, customDataIndex: %s, "+
			"params: %s)", visitorCode, customDataIndex, params)
	var p VisitorWarehouseAudienceOptParams
	if len(params) > 0 {
		p = params[0]
	}
	customData, err := c.warehouseManager.GetVisitorWarehouseAudience(
		visitorCode, p.WarehouseKey, customDataIndex, p.Timeout,
	)
	logging.Info(
		"RETURN: kameleoonClient.GetVisitorWarehouseAudienceWithOptParams(visitorCode: %s, customDataIndex: %s, "+
			"params: %s) -> (customData: %s, error: %s)", visitorCode, customDataIndex, params, customData, err)
	return customData, err
}

func (c *kameleoonClient) setUniqueIdentifier(visitorCode string, isUniqueIdentifier bool) {
	logging.Warning(
		"The 'isUniqueIdentifier' parameter is deprecated. Please, add 'UniqueIdentifier' to a visitor instead.")
	c.visitorManager.AddData(visitorCode, types.NewUniqueIdentifier(isUniqueIdentifier))
}

func (c *kameleoonClient) SetForcedVariation(
	visitorCode string, experimentId int, variationKey string, params ...SetForcedVariationOptParams,
) (err error) {
	logging.Info(
		"CALL: kameleoonClient.SetForcedVariation(visitorCode: %s, experimentId: %s, variationKey: %s, params: %s)",
		visitorCode, experimentId, variationKey, params,
	)
	defer func() {
		logging.Info(
			"RETURN: kameleoonClient.SetForcedVariation(visitorCode: %s, experimentId: %s, variationKey: %s, "+
				"params: %s) -> (error: %s)", visitorCode, experimentId, variationKey, params, err,
		)
	}()
	if err = utils.ValidateVisitorCode(visitorCode); err != nil {
		return
	}
	visitor := c.visitorManager.GetOrCreateVisitor(visitorCode)
	if variationKey != "" {
		ruleInfo, ruleInfoExists := c.dataManager.DataFile().GetRuleInfoByExpId(experimentId)
		if !ruleInfoExists {
			return errs.NewFeatureExperimentNotFound(experimentId)
		}
		var varByExp *types.VariationByExposition
		if varByExp, err = ruleInfo.Rule.GetVariationByKey(variationKey); err != nil {
			return
		}
		var p SetForcedVariationOptParams
		if len(params) > 0 {
			p = params[0]
		} else {
			p = NewSetForcedVariationOptParams()
		}
		forcedVariation := types.NewForcedExperimentVariation(ruleInfo.Rule, varByExp, p.forceTargeting)
		visitor.AddBaseData(true, forcedVariation)
	} else {
		visitor.ResetForcedVariation(experimentId)
	}
	return
}

func (c *kameleoonClient) EvaluateAudiences(visitorCode string) (err error) {
	logging.Info("CALL: kameleoonClient.EvaluateAudiences(visitorCode: %s)", visitorCode)
	defer func() {
		logging.Info("RETURN: kameleoonClient.EvaluateAudiences(visitorCode: %s) -> (error: %s)", visitorCode, err)
	}()
	if err = utils.ValidateVisitorCode(visitorCode); err != nil {
		return
	}
	var segments []types.BaseData
	for _, seg := range c.dataManager.DataFile().AudienceTrackingSegments() {
		if c.targetingManager.CheckTargeting(visitorCode, 0, seg) {
			segments = append(segments, types.NewTargetedSegment(seg.GetSegmentBase().ID))
		}
	}
	if len(segments) > 0 {
		c.visitorManager.GetOrCreateVisitor(visitorCode).AddBaseData(true, segments...)
	}
	c.trackingManager.AddVisitorCode(visitorCode)
	return
}

func (c *kameleoonClient) GetDataFile() types.DataFile {
	logging.Info("CALL: kameleoonClient.GetDataFile()")
	dataFile := c.dataManager.ExternalDataFile()
	logging.Info("RETURN: kameleoonClient.GetDataFile() -> (dataFile: %v)", dataFile)
	return *dataFile
}

type evaluatedExperiment struct {
	varByExp   *types.VariationByExposition
	experiment *types.Experiment
	ruleType   types.RuleType
}

func newEvaluatedExperimentFromVarByExpRule(
	varByExp *types.VariationByExposition, rule types.IRule,
) *evaluatedExperiment {
	return &evaluatedExperiment{
		varByExp:   varByExp,
		experiment: &rule.GetRuleBase().Experiment,
		ruleType:   rule.GetRuleBase().Type,
	}
}

func newEvaluatedExperimentFromForcedVariation(forcedVariation types.ForcedVariation) *evaluatedExperiment {
	if (forcedVariation.VarByExp() != nil) && (forcedVariation.Rule() != nil) {
		return newEvaluatedExperimentFromVarByExpRule(forcedVariation.VarByExp(), forcedVariation.Rule())
	}
	return nil
}

// func newEvaluatedExperimentFromForcedExperimentVariation(
// 	forcedVariation *types.ForcedExperimentVariation,
// ) *evaluatedExperiment {
// 	return newEvaluatedExperimentFromVarByExpRule(forcedVariation.VarByExp(), forcedVariation.Rule())
// }
