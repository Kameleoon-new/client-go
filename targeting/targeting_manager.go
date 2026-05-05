package targeting

import (
	"sync"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/managers/data"
	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/Kameleoon/client-go/v3/targeting/conditions"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
)

var (
	emptyVariations       = storage.NewDataMapStorageImpl[int, *types.AssignedVariation](new(sync.RWMutex), nil)
	emptyPersonalizations = storage.NewDataMapStorageImpl[int, *types.Personalization](new(sync.RWMutex), nil)
	emptyConversions      = storage.NewDataCollectionStorageImpl[*types.Conversion](new(sync.RWMutex), nil)
)

type TargetingManager interface {
	CheckTargeting(visitorCode string, campaignId int, segment types.Segment) bool
}

type targetingManager struct {
	visitorManager storage.VisitorManager
	dataManager    data.DataManager
}

func NewTargetingManager(dataManager data.DataManager, visitorManager storage.VisitorManager) TargetingManager {
	return &targetingManager{
		dataManager:    dataManager,
		visitorManager: visitorManager,
	}
}

func (tm *targetingManager) CheckTargeting(
	visitorCode string,
	campaignId int,
	segment types.Segment,
) bool {
	logging.Debug(
		"CALL: targetingManager.CheckTargeting(visitorCode: %s, campaignId: %s, segment: %s)",
		visitorCode, campaignId, segment,
	)
	visitor := tm.visitorManager.GetVisitor(visitorCode)
	targeted := segment == nil || segment.CheckTargeting(func(targetingType types.TargetingType) interface{} {
		return tm.getConditionData(targetingType, visitor, visitorCode, campaignId)
	})
	logging.Debug(
		"RETURN: targetingManager.CheckTargeting(visitorCode: %s, campaignId: %s, segment: %s) -> (targeted: %s)",
		visitorCode, campaignId, segment, targeted,
	)
	return targeted
}

func (tm *targetingManager) getConditionData(
	targetingType types.TargetingType,
	visitor storage.Visitor,
	visitorCode string,
	campaignId int,
) interface{} {
	logging.Debug(
		"CALL: targetingManager.getConditionData(targetingType: %s, visitor, visitorCode: %s, campaignId: %s)",
		targetingType, visitorCode, campaignId)
	var conditionData interface{}
	switch targetingType {
	case types.TargetingCustomDatum:
		if visitor != nil {
			conditionData = visitor.CustomData()
		}
	case types.TargetingBrowser:
		if visitor != nil {
			conditionData = visitor.Browser()
		}
	case types.TargetingDeviceType:
		if visitor != nil {
			conditionData = visitor.Device()
		}
	case types.TargetingApplicationVersion:
		if visitor != nil {
			conditionData = visitor.ApplicationVersion()
		}
	case types.TargetingPageTitle:
		fallthrough
	case types.TargetingPageUrl:
		fallthrough
	case types.TargetingPageViews:
		fallthrough
	case types.TargetingPreviousPage:
		if visitor != nil {
			conditionData = visitor.PageViewVisits()
		}
	case types.TargetingConversions:
		conditionData = conditions.TargetingDataConversionCondition{
			VisitorVisits: getVisitorVisits(visitor),
			Conversions:   getConversions(visitor),
		}
	case types.TargetingVisitorCode:
		conditionData = visitorCode
	case types.TargetingSDKLanguage:
		conditionData = &types.TargetedDataSdk{Language: utils.SdkName, Version: utils.SdkVersion}
	case types.TargetingTargetFeatureFlag:
		conditionData = conditions.TargetingDataTargetFeatureFlagCondition{
			VisitorVisits:    getVisitorVisits(visitor),
			DataFile:         tm.dataManager.DataFile(),
			VariationStorage: getVariations(visitor),
		}
	case types.TargetingTargetExperiment:
		conditionData = conditions.TargetingDataTargetExperimentCondition{
			VisitorVisits:    getVisitorVisits(visitor),
			VariationStorage: getVariations(visitor),
		}
	case types.TargetingTargetPersonalization:
		conditionData = conditions.TargetingDataTargetPersonalizationCondition{
			VisitorVisits:    getVisitorVisits(visitor),
			Personalizations: getPersonalizations(visitor),
		}
	case types.TargetingExclusiveExperiment:
		if campaignId > 0 {
			conditionData = conditions.TargetingDataExclusiveExperiment{
				VisitorVisits:       getVisitorVisits(visitor),
				CurrentExperimentId: campaignId,
				Variations:          getVariations(visitor),
				Personalizations:    getPersonalizations(visitor),
			}
		}
	case types.TargetingCookie:
		if visitor != nil {
			conditionData = visitor.Cookie()
		}
	case types.TargetingGeolocation:
		if visitor != nil {
			conditionData = visitor.Geolocation()
		}
	case types.TargetingOperatingSystem:
		if visitor != nil {
			conditionData = visitor.OperatingSystem()
		}
	case types.TargetingSegment:
		conditionData = conditions.TargetingDataSegmentCondition{
			DataFile: tm.dataManager.DataFile(),
			TargetingDataGetter: func(targetingType types.TargetingType) interface{} {
				return tm.getConditionData(targetingType, visitor, visitorCode, campaignId)
			},
		}
	case types.TargetingFirstVisit:
		fallthrough
	case types.TargetingLastVisit:
		fallthrough
	case types.TargetingVisits:
		fallthrough
	case types.TargetingNewVisitors:
		if visitor != nil {
			conditionData = visitor.VisitorVisits()
		} else {
			conditionData = (*types.VisitorVisits)(nil)
		}
	case types.TargetingSameDayVisits:
		if visitor != nil {
			conditionData = conditions.TargetingDataVisitNumberToday{
				CurrentVisitTimeStarted: visitor.TimeStarted(),
				VisitorVisits:           visitor.VisitorVisits(),
			}
		} else {
			conditionData = conditions.TargetingDataVisitNumberToday{CurrentVisitTimeStarted: time.Now()}
		}
	case types.TargetingHeatSlice:
		if visitor != nil {
			conditionData = visitor.KcsHeat()
		}
	}
	logging.Debug(
		"RETURN: targetingManager.getConditionData(targetingType: %s, visitor, visitorCode: %s, campaignId: %s) "+
			"-> (conditionData: %s)", targetingType, visitorCode, campaignId, conditionData)
	return conditionData
}

func getVisitorVisits(visitor storage.Visitor) *types.VisitorVisits {
	if visitor != nil {
		return visitor.VisitorVisits()
	}
	return nil
}

func getConversions(visitor storage.Visitor) storage.DataCollectionStorage[*types.Conversion] {
	if visitor != nil {
		return visitor.Conversions()
	}
	return emptyConversions
}

func getVariations(visitor storage.Visitor) storage.DataMapStorage[int, *types.AssignedVariation] {
	if visitor != nil {
		return visitor.Variations()
	}
	return emptyVariations
}

func getPersonalizations(visitor storage.Visitor) storage.DataMapStorage[int, *types.Personalization] {
	if visitor != nil {
		return visitor.Personalizations()
	}
	return emptyPersonalizations
}
