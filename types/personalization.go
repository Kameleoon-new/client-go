package types

import (
	"fmt"
	"time"
)

type Personalization struct {
	id             int
	variationId    int
	assignmentTime time.Time
}

// Note: This is intended for internal use only and is not part of the stable API.
func NewPersonalization(id int, variationId int, assignmentTime time.Time) *Personalization {
	return &Personalization{id: id, variationId: variationId, assignmentTime: assignmentTime}
}

func (p *Personalization) Id() int {
	return p.id
}

func (p *Personalization) VariationId() int {
	return p.variationId
}

func (p *Personalization) AssignmentTime() time.Time {
	return p.assignmentTime
}

func (*Personalization) DataType() DataType {
	return DataTypePersonalization
}

func (p Personalization) String() string {
	return fmt.Sprintf(
		"Personalization{id:%d,variationId:%d,assignmentTime:%v}",
		p.id,
		p.variationId,
		p.assignmentTime,
	)
}
