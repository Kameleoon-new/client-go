package remotedata

import (
	"encoding/json"
	"time"

	"github.com/Kameleoon/client-go/v3/types"
)

type remoteVisitorData struct {
	filter           types.RemoteVisitorDataFilter
	customDataDict   map[int]types.ICustomData
	pageViewVisits   map[string]types.PageViewVisit
	conversions      []*types.Conversion
	experiments      map[int]*types.AssignedVariation
	personalizations map[int]*types.Personalization
	device           *types.Device
	browser          *types.Browser
	operatingSystem  *types.OperatingSystem
	geolocation      *types.Geolocation
	visitNumber      int
	visitorVisits    *types.VisitorVisits
	kcsHeat          *types.KcsHeat
	cbs              *types.CBScores
	visitorCode      string
}

func newRemoteVisitorData(filter types.RemoteVisitorDataFilter) *remoteVisitorData {
	return &remoteVisitorData{filter: filter}
}

func (rvd *remoteVisitorData) MarkVisitorDataAsSent(customDataInfo *types.CustomDataInfo) {
	for id, cd := range rvd.customDataDict {
		if (customDataInfo == nil) || !customDataInfo.IsVisitorScope(id) {
			cd.MarkAsSent()
		}
	}
	for _, pvv := range rvd.pageViewVisits {
		pvv.PageView.MarkAsSent()
	}
	for _, c := range rvd.conversions {
		c.MarkAsSent()
	}
	for _, av := range rvd.experiments {
		av.MarkAsSent()
	}
	if rvd.device != nil {
		rvd.device.MarkAsSent()
	}
	if rvd.browser != nil {
		rvd.browser.MarkAsSent()
	}
	if rvd.operatingSystem != nil {
		rvd.operatingSystem.MarkAsSent()
	}
	if rvd.geolocation != nil {
		rvd.geolocation.MarkAsSent()
	}
}

func (rvd *remoteVisitorData) CollectVisitorDataToReturn() []types.Data {
	dataList := make([]types.Data, 0, len(rvd.customDataDict)+len(rvd.pageViewVisits)+len(rvd.conversions)+4)
	for _, cd := range rvd.customDataDict {
		dataList = append(dataList, cd)
	}
	for _, pvv := range rvd.pageViewVisits {
		dataList = append(dataList, pvv.PageView)
	}
	for _, c := range rvd.conversions {
		dataList = append(dataList, c)
	}
	if rvd.device != nil {
		dataList = append(dataList, rvd.device)
	}
	if rvd.browser != nil {
		dataList = append(dataList, rvd.browser)
	}
	if rvd.operatingSystem != nil {
		dataList = append(dataList, rvd.operatingSystem)
	}
	if rvd.geolocation != nil {
		dataList = append(dataList, rvd.geolocation)
	}
	return dataList
}

func (rvd *remoteVisitorData) CollectDataToAdd() []types.BaseData {
	dataList := make([]types.BaseData, 0,
		len(rvd.customDataDict)+len(rvd.pageViewVisits)+len(rvd.conversions)+len(rvd.experiments)+5)
	for _, cd := range rvd.customDataDict {
		dataList = append(dataList, cd)
	}
	for _, pvv := range rvd.pageViewVisits {
		dataList = append(dataList, pvv)
	}
	for _, c := range rvd.conversions {
		dataList = append(dataList, c)
	}
	for _, av := range rvd.experiments {
		dataList = append(dataList, av)
	}
	for _, p := range rvd.personalizations {
		dataList = append(dataList, p)
	}
	if rvd.device != nil {
		dataList = append(dataList, rvd.device)
	}
	if rvd.browser != nil {
		dataList = append(dataList, rvd.browser)
	}
	if rvd.operatingSystem != nil {
		dataList = append(dataList, rvd.operatingSystem)
	}
	if rvd.geolocation != nil {
		dataList = append(dataList, rvd.geolocation)
	}
	if rvd.visitorVisits != nil {
		dataList = append(dataList, rvd.visitorVisits)
	}
	if rvd.kcsHeat != nil {
		dataList = append(dataList, rvd.kcsHeat)
	}
	if rvd.cbs != nil {
		dataList = append(dataList, rvd.cbs)
	}
	return dataList
}

func (rvd *remoteVisitorData) UnmarshalJSON(data []byte) error {
	m := new(remoteVisitorDataModel)
	if err := json.Unmarshal(data, m); err != nil {
		return err
	}
	rvd.customDataDict = make(map[int]types.ICustomData)
	rvd.pageViewVisits = make(map[string]types.PageViewVisit)
	rvd.conversions = []*types.Conversion{}
	rvd.experiments = make(map[int]*types.AssignedVariation)
	rvd.personalizations = make(map[int]*types.Personalization)
	rvd.device = nil
	rvd.browser = nil
	rvd.operatingSystem = nil
	rvd.geolocation = nil
	rvd.parseCurrentVisit(m)
	rvd.parsePreviousVisits(m)
	rvd.parseKcsHeat(m)
	rvd.parseCBScores(m)
	return nil
}

func (rvd *remoteVisitorData) parseCurrentVisit(m *remoteVisitorDataModel) {
	if m.CurrentVisit != nil {
		rvd.parseVisit(m.CurrentVisit, 0)
	}
}

func (rvd *remoteVisitorData) parsePreviousVisits(m *remoteVisitorDataModel) {
	prevVisits := make([]types.Visit, 0, len(m.PreviousVisits))
	for i, prevVisit := range m.PreviousVisits {
		if prevVisit != nil {
			prevVisits = append(prevVisits, types.NewVisit(prevVisit.TimeStarted, prevVisit.TimeLastEvent))
			rvd.parseVisit(prevVisit, i+1)
		}
	}
	if len(prevVisits) > 0 {
		rvd.visitorVisits = types.NewVisitorVisits(prevVisits, rvd.visitNumber)
	} else {
		rvd.visitorVisits = nil
	}
}

func (rvd *remoteVisitorData) parseVisit(v *visitModel /*non-nil*/, visitOffset int) {
	if rvd.visitorCode == "" {
		rvd.visitorCode = v.VisitorCode
	}
	rvd.parseCustomData(v.CustomDataEvents)
	rvd.parsePages(v.PageEvents)
	rvd.parseExperiments(v.ExperimentEvents)
	rvd.parsePersonalizations(v.PersonalizationEvents)
	rvd.parseConversions(v.ConversionEvents)
	rvd.parseGeolocation(v.GeolocationEvents)
	rvd.parseStaticData(v.StaticDataEvent, visitOffset)
}

func (rvd *remoteVisitorData) parseCustomData(customDataEvents []*dataEventModel[customDataModel]) {
	for i := len(customDataEvents) - 1; i >= 0; i-- {
		event := customDataEvents[i]
		if (event == nil) || (event.Data == nil) {
			continue
		}
		if _, contains := rvd.customDataDict[event.Data.Index]; !contains {
			var values []string
			if event.Data.ValuesCountMap != nil {
				values = make([]string, 0, len(event.Data.ValuesCountMap))
				for value := range event.Data.ValuesCountMap {
					values = append(values, value)
				}
			}
			rvd.customDataDict[event.Data.Index] = types.NewCustomData(event.Data.Index, values...)
		}
	}
}

func (rvd *remoteVisitorData) parsePages(pageEvents []*dataEventModel[pageDataModel]) {
	for i := len(pageEvents) - 1; i >= 0; i-- {
		event := pageEvents[i]
		if (event == nil) || (event.Data == nil) || (event.Data.Href == "") {
			continue
		}
		var pageViewVisit types.PageViewVisit
		var contains bool
		if pageViewVisit, contains = rvd.pageViewVisits[event.Data.Href]; contains {
			pageViewVisit.Count++
		} else {
			pageViewVisit.PageView = types.NewPageViewWithTitle(event.Data.Href, event.Data.Title)
			pageViewVisit.Count = 1
			pageViewVisit.LastTimestamp = event.Time
		}
		rvd.pageViewVisits[event.Data.Href] = pageViewVisit
	}
}

func (rvd *remoteVisitorData) parseExperiments(experimentEvents []*dataEventModel[experimentDataModel]) {
	for i := len(experimentEvents) - 1; i >= 0; i-- {
		event := experimentEvents[i]
		if (event == nil) || (event.Data == nil) {
			continue
		}
		if _, contains := rvd.experiments[event.Data.ExperimentId]; !contains {
			rvd.experiments[event.Data.ExperimentId] = types.NewAssignedVariationWithTime(
				event.Data.ExperimentId,
				event.Data.VariationId,
				types.RuleTypeUnknown,
				time.Unix(0, event.Time*int64(time.Millisecond)),
			)
		}
	}
}

func (rvd *remoteVisitorData) parsePersonalizations(personalizationEvents []*dataEventModel[personalizationDataModel]) {
	for i := len(personalizationEvents) - 1; i >= 0; i-- {
		event := personalizationEvents[i]
		if (event == nil) || (event.Data == nil) {
			continue
		}
		if _, contains := rvd.personalizations[event.Data.Id]; !contains {
			rvd.personalizations[event.Data.Id] = types.NewPersonalization(
				event.Data.Id,
				event.Data.VariationId,
				time.Unix(0, event.Time*int64(time.Millisecond)),
			)
		}
	}
}

func (rvd *remoteVisitorData) parseConversions(conversionEvents []*dataEventModel[conversionDataModel]) {
	for _, event := range conversionEvents {
		if (event == nil) || (event.Data == nil) {
			continue
		}
		c := types.InternalNewConversion(
			event.Data.GoalId,
			types.ConversionOptParams{
				Revenue:  event.Data.Revenue,
				Negative: event.Data.Negative,
			},
			time.Unix(0, event.Time*int64(time.Millisecond)),
		)
		rvd.conversions = append(rvd.conversions, c)
	}
}

func (rvd *remoteVisitorData) parseGeolocation(geolocationEvents []*dataEventModel[geolocationDataModel]) {
	if (rvd.geolocation != nil) || (len(geolocationEvents) == 0) {
		return
	}
	event := geolocationEvents[len(geolocationEvents)-1]
	if (event == nil) || (event.Data == nil) {
		return
	}
	rvd.geolocation = types.NewGeolocation(event.Data.Country, event.Data.Region, event.Data.City)
}

func (rvd *remoteVisitorData) parseStaticData(staticDataEvent *dataEventModel[staticDataModel], visitOffset int) {
	if (staticDataEvent == nil) || (staticDataEvent.Data == nil) {
		return
	}
	if (rvd.visitNumber == 0) && (staticDataEvent.Data.VisitNumber != nil) {
		rvd.visitNumber = *staticDataEvent.Data.VisitNumber + visitOffset
	}
	if rvd.filter.Device && (rvd.device == nil) {
		rvd.device = types.NewDevice(types.DeviceType(staticDataEvent.Data.DeviceType))
	}
	if rvd.filter.Browser && (rvd.browser == nil) {
		if browserType, ok := types.ParseBrowserType(staticDataEvent.Data.Browser); ok {
			rvd.browser = types.NewBrowser(browserType, staticDataEvent.Data.BrowserVersion)
		}
	}
	if rvd.filter.OperatingSystem && (rvd.operatingSystem == nil) {
		if osType, ok := types.ParseOperatingSystemType(staticDataEvent.Data.OsType); ok {
			rvd.operatingSystem = types.NewOperatingSystem(osType)
		}
	}
}

func (rvd *remoteVisitorData) parseKcsHeat(m *remoteVisitorDataModel) {
	if m.Kcs != nil {
		rvd.kcsHeat = types.NewKcsHeat(m.Kcs)
	} else {
		rvd.kcsHeat = nil
	}
}

func (rvd *remoteVisitorData) parseCBScores(m *remoteVisitorDataModel) {
	if m.Cbs != nil {
		cbsMap := make(map[int][]types.ScoredVarId)
		for experimentId, scoredVarEntries := range m.Cbs {
			entries := make([]types.ScoredVarId, 0, len(scoredVarEntries))
			for varId, score := range scoredVarEntries {
				entries = append(entries, types.ScoredVarId{VariationId: varId, Score: score})
			}
			cbsMap[experimentId] = entries
		}
		rvd.cbs = types.NewCBScores(cbsMap)
	} else {
		rvd.cbs = nil
	}
}

type remoteVisitorDataModel struct {
	CurrentVisit   *visitModel             `json:"currentVisit"`
	PreviousVisits []*visitModel           `json:"previousVisits"`
	Kcs            map[int]map[int]float64 `json:"kcs"`
	Cbs            map[int]map[int]float64 `json:"cbs"`
}

type visitModel struct {
	TimeStarted           int64                                       `json:"timeStarted"`
	TimeLastEvent         int64                                       `json:"timeLastEvent"`
	VisitorCode           string                                      `json:"visitorCode"`
	CustomDataEvents      []*dataEventModel[customDataModel]          `json:"customDataEvents"`
	PageEvents            []*dataEventModel[pageDataModel]            `json:"pageEvents"`
	ExperimentEvents      []*dataEventModel[experimentDataModel]      `json:"experimentEvents"`
	PersonalizationEvents []*dataEventModel[personalizationDataModel] `json:"personalizationEvents"`
	ConversionEvents      []*dataEventModel[conversionDataModel]      `json:"conversionEvents"`
	GeolocationEvents     []*dataEventModel[geolocationDataModel]     `json:"geolocationEvents"`
	StaticDataEvent       *dataEventModel[staticDataModel]            `json:"staticDataEvent"`
}

type dataEventModel[D any] struct {
	Data *D    `json:"data"`
	Time int64 `json:"time"`
}

type customDataModel struct {
	Index          int            `json:"index"`
	ValuesCountMap map[string]int `json:"valuesCountMap"`
}

type pageDataModel struct {
	Href  string `json:"href"`
	Title string `json:"title"`
}

type experimentDataModel struct {
	ExperimentId int `json:"id"`
	VariationId  int `json:"variationId"`
}

type personalizationDataModel struct {
	Id          int `json:"id"`
	VariationId int `json:"variationId"`
}

type conversionDataModel struct {
	GoalId   int     `json:"goalId"`
	Revenue  float64 `json:"revenue"`
	Negative bool    `json:"negative"`
}

type geolocationDataModel struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
}

type staticDataModel struct {
	VisitNumber    *int    `json:"visitNumber,omitempty"`
	DeviceType     string  `json:"deviceType"`
	Browser        string  `json:"browser"`
	BrowserVersion float32 `json:"browserVersion"`
	OsType         string  `json:"os"`
}
