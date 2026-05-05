package conditions

import (
	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/Kameleoon/client-go/v3/types"
)

type TargetPersonalizationCondition struct {
	types.TargetingConditionBase
	visitorScopeCondition
	personalizationId int
}

func NewTargetPersonalizationCondition(c types.TargetingCondition) *TargetPersonalizationCondition {
	return &TargetPersonalizationCondition{
		TargetingConditionBase: types.TargetingConditionBase{
			Type:    c.Type,
			Include: true,
		},
		visitorScopeCondition: newVisitorScopeCondition(c, VisitScopeCurrentVisit),
		personalizationId:     c.PersonalizationId,
	}
}

func (c *TargetPersonalizationCondition) CheckTargeting(targetData interface{}) bool {
	targetingData, ok := targetData.(TargetingDataTargetPersonalizationCondition)
	if !ok || (targetingData.Personalizations == nil) || (c.personalizationId == types.UndefinedPersonalizationId) {
		return false
	}
	personalization := targetingData.Personalizations.Get(c.personalizationId)
	return (personalization != nil) &&
		(personalization.AssignmentTime().UnixMilli() >= c.assignmentThresholdMillis(targetingData.VisitorVisits))
}

type TargetingDataTargetPersonalizationCondition struct {
	VisitorVisits    *types.VisitorVisits
	Personalizations storage.DataMapStorage[int, *types.Personalization]
}
