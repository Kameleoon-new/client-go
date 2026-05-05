package conditions

import (
	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/Kameleoon/client-go/v3/types"
)

type TargetExperimentCondition struct {
	types.TargetingConditionBase
	visitorScopeCondition
	variationId        int
	experimentId       int
	variationMatchType types.OperatorType
}

func NewTargetExperimentCondition(c types.TargetingCondition) *TargetExperimentCondition {
	return &TargetExperimentCondition{
		TargetingConditionBase: types.TargetingConditionBase{
			Type:    c.Type,
			Include: true,
		},
		visitorScopeCondition: newVisitorScopeCondition(c, VisitScopeCurrentVisit),
		variationId:           c.VariationId,
		experimentId:          c.ExperimentId,
		variationMatchType:    c.VariationMatchType,
	}
}

func (c *TargetExperimentCondition) CheckTargeting(targetData interface{}) bool {
	targetingData, ok := targetData.(TargetingDataTargetExperimentCondition)
	if !ok || (targetingData.VariationStorage == nil) {
		return false
	}
	variation := targetingData.VariationStorage.Get(c.experimentId)
	if (variation != nil) &&
		(variation.AssignmentTime().UnixMilli() < c.assignmentThresholdMillis(targetingData.VisitorVisits)) {
		variation = nil
	}
	switch c.variationMatchType {
	case types.OperatorAny:
		return variation != nil
	case types.OperatorExact:
		return (variation != nil) && (variation.VariationId() == c.variationId)
	}
	logging.Error("Unexpected variation match type for %s condition: %s", c.Type, c.variationMatchType)
	return false
}

type TargetingDataTargetExperimentCondition struct {
	VisitorVisits    *types.VisitorVisits
	VariationStorage storage.DataMapStorage[int, *types.AssignedVariation]
}
