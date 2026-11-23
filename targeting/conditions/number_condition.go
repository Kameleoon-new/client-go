package conditions

import (
	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
	"golang.org/x/exp/constraints"
)

type NumberCondition[T constraints.Ordered] struct {
	types.TargetingConditionBase
	Value     T                  `json:"value"`
	MatchType types.OperatorType `json:"matchType"`
}

func (c *NumberCondition[T]) CheckTargeting(targetData interface{}) bool {
	value, ok := targetData.(T)
	return ok && c.checkTargeting(value)
}

func (c *NumberCondition[T]) checkTargeting(value T) bool {
	targeted, isComparison := compare(c.MatchType, value, c.Value)
	if !isComparison {
		logging.Error("Unexpected comparing operation for %s condition: %s", c.TargetingConditionBase.Type, c.MatchType)
	}
	return targeted
}

func (c NumberCondition[T]) String() string {
	return utils.JsonToString(c)
}
