package conditions

import (
	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
)

func NewConversionCondition(c types.TargetingCondition) *ConversionCondition {
	return &ConversionCondition{
		TargetingConditionBase: types.TargetingConditionBase{
			Type:    c.Type,
			Include: c.Include,
		},
		visitorScopeCondition: newVisitorScopeCondition(c, VisitScopeVisitor),
		GoalId:                c.GoalId,
	}
}

type ConversionCondition struct {
	types.TargetingConditionBase
	visitorScopeCondition
	GoalId int `json:"goalId"`
}

func (c *ConversionCondition) CheckTargeting(targetData interface{}) bool {
	targetingData, ok := targetData.(TargetingDataConversionCondition)
	if !ok || (targetingData.Conversions == nil) || (targetingData.Conversions.Len() == 0) {
		return false
	}
	assignmentThresholdMillis := c.assignmentThresholdMillis(targetingData.VisitorVisits)
	targeted := false
	targetingData.Conversions.Enumerate(func(conversion *types.Conversion) bool {
		targeted = ((c.GoalId == types.UndefinedGoalId) || (c.GoalId == conversion.GoalId())) &&
			(conversion.AssignmentTime().UnixMilli() >= assignmentThresholdMillis)
		return !targeted
	})
	return targeted
}

func (c ConversionCondition) String() string {
	return utils.JsonToString(c)
}

type TargetingDataConversionCondition struct {
	VisitorVisits *types.VisitorVisits
	Conversions   storage.DataCollectionStorage[*types.Conversion]
}
