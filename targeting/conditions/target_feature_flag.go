package conditions

import (
	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
)

type TargetFeatureFlagCondition struct {
	types.TargetingConditionBase
	visitorScopeCondition
	FeatureFlagId         int    `json:"featureFlagId,omitempty"`
	ConditionVariationKey string `json:"variationKey,omitempty"`
	ConditionRuleId       int    `json:"ruleId,omitempty"`
}

func NewTargetFeatureFlagCondition(c types.TargetingCondition) *TargetFeatureFlagCondition {
	return &TargetFeatureFlagCondition{
		TargetingConditionBase: types.TargetingConditionBase{
			Type:    c.Type,
			Include: true,
		},
		visitorScopeCondition: newVisitorScopeCondition(c, VisitScopeCurrentVisit),
		FeatureFlagId:         c.FeatureFlagId,
		ConditionVariationKey: c.VariationKey,
		ConditionRuleId:       c.RuleId,
	}
}

func (c *TargetFeatureFlagCondition) CheckTargeting(targetData interface{}) bool {
	td, ok := targetData.(TargetingDataTargetFeatureFlagCondition)
	if !ok || td.DataFile == nil || td.VariationStorage == nil || td.VariationStorage.Len() == 0 {
		return false
	}
	featureFlag := td.DataFile.GetFeatureFlagById(c.FeatureFlagId)
	if featureFlag == nil {
		return false
	}
	assignmentThresholdMillis := c.assignmentThresholdMillis(td.VisitorVisits)
	for _, rule := range featureFlag.GetRules() {
		if rule == nil || rule.GetRuleBase() == nil {
			continue
		}
		base := rule.GetRuleBase()
		if c.ConditionRuleId > 0 && base.Id != c.ConditionRuleId {
			continue
		}
		assignedVariation := td.VariationStorage.Get(base.ExperimentId)
		if assignedVariation == nil || assignedVariation.AssignmentTime().UnixMilli() < assignmentThresholdMillis {
			continue
		}
		if c.matchesVariationKey(td, assignedVariation) {
			return true
		}
	}
	return false
}

func (c *TargetFeatureFlagCondition) matchesVariationKey(
	targetingData TargetingDataTargetFeatureFlagCondition, variation *types.AssignedVariation,
) bool {
	if c.ConditionVariationKey == "" {
		return true
	}
	vbe := targetingData.DataFile.GetVariation(variation.VariationId())
	return (vbe != nil) && (vbe.VariationKey == c.ConditionVariationKey)
}

func (c TargetFeatureFlagCondition) String() string {
	return utils.JsonToString(c)
}

type TargetingDataTargetFeatureFlagCondition struct {
	VisitorVisits    *types.VisitorVisits
	DataFile         types.IDataFile
	VariationStorage storage.DataMapStorage[int, *types.AssignedVariation]
}
