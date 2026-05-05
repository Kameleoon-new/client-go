package types

import (
	"fmt"
	"time"

	"github.com/Kameleoon/client-go/v3/utils"
)

const assignedVariationEventType = "experiment"

type AssignedVariation struct {
	duplicationUnsafeSendableBase
	experimentId   int
	variationId    int
	ruleType       RuleType
	assignmentTime time.Time
}

// Note: This is intended for internal use only and is not part of the stable API.
func NewAssignedVariation(experimentId int, variationId int, ruleType RuleType) *AssignedVariation {
	return NewAssignedVariationWithTime(experimentId, variationId, ruleType, time.Now())
}

// Note: This is intended for internal use only and is not part of the stable API.
func NewAssignedVariationWithTime(
	experimentId int, variationId int, ruleType RuleType, assignmentTime time.Time,
) *AssignedVariation {
	return &AssignedVariation{
		experimentId:   experimentId,
		variationId:    variationId,
		ruleType:       ruleType,
		assignmentTime: assignmentTime,
	}
}

func (av AssignedVariation) String() string {
	return fmt.Sprintf(
		"AssignedVariation{experimentId:%v,variationId:%v,assignmentTime:%v,ruleType:%v}",
		av.experimentId,
		av.variationId,
		av.assignmentTime,
		av.ruleType,
	)
}

func (av *AssignedVariation) ExperimentId() int {
	return av.experimentId
}

func (av *AssignedVariation) VariationId() int {
	return av.variationId
}

func (av *AssignedVariation) RuleType() RuleType {
	return av.ruleType
}

func (av *AssignedVariation) AssignmentTime() time.Time {
	return av.assignmentTime
}

func (av *AssignedVariation) IsValid(respoolTime int) bool {
	return (respoolTime == 0) || (av.assignmentTime.UnixMilli() >= int64(respoolTime))
}

func (av *AssignedVariation) QueryEncode() string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPEventType, assignedVariationEventType)
	qb.Append(utils.QPExperimentId, fmt.Sprint(av.experimentId))
	qb.Append(utils.QPVariationId, fmt.Sprint(av.variationId))
	qb.Append(utils.QPNonce, av.Nonce())
	return qb.String()
}

func (av *AssignedVariation) DataType() DataType {
	return DataTypeAssignedVariation
}
