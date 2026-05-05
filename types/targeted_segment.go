package types

import (
	"fmt"
	"strconv"

	"github.com/Kameleoon/client-go/v3/utils"
)

const targetedSegmentEventType = "targetingSegment"

type TargetedSegment struct {
	duplicationUnsafeSendableBase
	id int
}

func NewTargetedSegment(id int) *TargetedSegment {
	return &TargetedSegment{id: id}
}

func (ts *TargetedSegment) Id() int {
	return ts.id
}

func (ts *TargetedSegment) QueryEncode() string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPEventType, targetedSegmentEventType)
	qb.Append(utils.QPSegmentId, strconv.Itoa(ts.id))
	qb.Append(utils.QPNonce, ts.Nonce())
	return qb.String()
}

func (*TargetedSegment) DataType() DataType {
	return DataTypeTargetedSegment
}

func (ts TargetedSegment) String() string {
	return fmt.Sprintf("TargetedSegment{id:%d}", ts.id)
}
