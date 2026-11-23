package conditions

import (
	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
)

func NewVersionCondition(c types.TargetingCondition) *VersionCondition {
	return &VersionCondition{
		TargetingConditionBase: types.TargetingConditionBase{
			Type:    c.Type,
			Include: c.Include,
		},
		Version:          c.Version,
		VersionMatchType: c.VersionMatchType,
	}
}

type VersionCondition struct {
	types.TargetingConditionBase
	Version          string             `json:"version"`
	VersionMatchType types.OperatorType `json:"versionMatchType,omitempty"`
}

func (c *VersionCondition) CheckTargeting(targetData interface{}) bool {
	av, ok := targetData.(*types.ApplicationVersion)
	return ok && av != nil && c.CompareWithVersion(av.Version)
}

func (c *VersionCondition) CompareWithVersion(targetVersion string) bool {
	condition, err := utils.NewVersionFromString(c.Version)
	if err != nil {
		logging.Error("Failed to parse version '%s' for '%s' condition", c.Version, c.Type)
		return false
	}
	target, err := utils.NewVersionFromString(targetVersion)
	if err != nil {
		logging.Error("Failed to parse version '%s' for target in '%s' condition", targetVersion, c.Type)
		return false
	}

	// The sign of comparing the versions is compared with zero
	targeted, isComparison := compare(c.VersionMatchType, target.CompareTo(condition), 0)
	if !isComparison {
		logging.Error("Unexpected comparing operation for '%s' condition: %s", c.Type, c.VersionMatchType)
	}
	return targeted
}

func (c VersionCondition) String() string {
	return utils.JsonToString(c)
}
