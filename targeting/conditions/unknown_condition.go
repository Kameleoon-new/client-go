package conditions

import (
	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/types"
)

type UnknownCondition struct {
	types.TargetingConditionBase
}

func NewUnknownCondition(c types.TargetingCondition) *UnknownCondition {
	return &UnknownCondition{
		TargetingConditionBase: types.TargetingConditionBase{
			Type:    c.Type,
			Include: c.Include,
		},
	}
}

func (c *UnknownCondition) CheckTargeting(targetData interface{}) bool {
	logging.Warning("Condition of unknown type '%s' evaluated as true", c.Type)
	return true
}
