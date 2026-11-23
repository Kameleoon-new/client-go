package conditions

import (
	"github.com/Kameleoon/client-go/v3/types"
	"golang.org/x/exp/constraints"
)

// compare checks `value <operator> conditionValue`.
//
// isComparison is false if the operator is not a comparison one; the result is false then.
// A NaN value never matches.
func compare[T constraints.Ordered](
	operator types.OperatorType, value T, conditionValue T,
) (targeted bool, isComparison bool) {
	switch operator {
	case types.OperatorLower:
		return value < conditionValue, true
	case types.OperatorGreater:
		return value > conditionValue, true
	case types.OperatorEqual:
		return value == conditionValue, true
	case types.OperatorLowerOrEqual:
		return value <= conditionValue, true
	case types.OperatorGreaterOrEqual:
		return value >= conditionValue, true
	default:
		return false, false
	}
}

func isComparison(operator types.OperatorType) bool {
	switch operator {
	case types.OperatorLower, types.OperatorGreater, types.OperatorEqual,
		types.OperatorLowerOrEqual, types.OperatorGreaterOrEqual:
		return true
	default:
		return false
	}
}
