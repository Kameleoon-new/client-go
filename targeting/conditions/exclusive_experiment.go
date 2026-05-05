package conditions

import (
	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
)

const (
	CampaignTypeExperiment      = "EXPERIMENT"
	CampaignTypePersonalization = "PERSONALIZATION"
	CampaignTypeAny             = "ANY"
)

type ExclusiveExperimentCondition struct {
	types.TargetingConditionBase
	visitorScopeCondition
	campaignType string
}

func NewExclusiveExperimentCondition(c types.TargetingCondition) *ExclusiveExperimentCondition {
	return &ExclusiveExperimentCondition{
		TargetingConditionBase: types.TargetingConditionBase{
			Type:    c.Type,
			Include: true,
		},
		visitorScopeCondition: newVisitorScopeCondition(c, VisitScopeVisitor),
		campaignType:          c.CampaignType,
	}
}

func (c *ExclusiveExperimentCondition) CheckTargeting(targetData interface{}) bool {
	if targetingData, ok := targetData.(TargetingDataExclusiveExperiment); ok {
		assignmentThresholdMillis := c.assignmentThresholdMillis(targetingData.VisitorVisits)
		switch c.campaignType {
		case CampaignTypeExperiment:
			return c.checkExperiment(targetingData, assignmentThresholdMillis)
		case CampaignTypePersonalization:
			return c.checkPersonalization(targetingData, assignmentThresholdMillis)
		case CampaignTypeAny:
			return c.checkPersonalization(targetingData, assignmentThresholdMillis) &&
				c.checkExperiment(targetingData, assignmentThresholdMillis)
		}
		logging.Error("Unexpected campaign type for %s condition: %s", c.Type, c.campaignType)
	}
	return false
}

func (*ExclusiveExperimentCondition) checkExperiment(
	td TargetingDataExclusiveExperiment, assignmentThresholdMillis int64,
) bool {
	if td.Variations == nil {
		return true
	}
	targeted := true
	td.Variations.Enumerate(func(assignedVariation *types.AssignedVariation) bool {
		if (assignedVariation.ExperimentId() != td.CurrentExperimentId) &&
			(assignedVariation.AssignmentTime().UnixMilli() >= assignmentThresholdMillis) {
			targeted = false
		}
		return targeted
	})
	return targeted
}

func (*ExclusiveExperimentCondition) checkPersonalization(
	td TargetingDataExclusiveExperiment, assignmentThresholdMillis int64,
) bool {
	if td.Personalizations == nil {
		return true
	}
	targeted := true
	td.Personalizations.Enumerate(func(personalization *types.Personalization) bool {
		if personalization.AssignmentTime().UnixMilli() >= assignmentThresholdMillis {
			targeted = false
		}
		return targeted
	})
	return targeted
}

func (c ExclusiveExperimentCondition) String() string {
	return utils.JsonToString(c)
}

type TargetingDataExclusiveExperiment struct {
	VisitorVisits       *types.VisitorVisits
	CurrentExperimentId int
	Variations          storage.DataMapStorage[int, *types.AssignedVariation]
	Personalizations    storage.DataMapStorage[int, *types.Personalization]
}
