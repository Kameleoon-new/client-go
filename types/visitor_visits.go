package types

import (
	"fmt"
	"strconv"

	"github.com/Kameleoon/client-go/v3/utils"
)

const visitorVisitsEventType = "staticData"

type VisitorVisits struct {
	duplicationUnsafeSendableBase
	visitNumber            int
	prevVisits             []Visit
	timeStarted            int64
	timeSincePreviousVisit int64
}

func NewVisitorVisits(prevVisits []Visit, visitNumber ...int) *VisitorVisits {
	var visitNumberValue int
	if (len(visitNumber) > 0) && (visitNumber[0] > len(prevVisits)) {
		visitNumberValue = visitNumber[0]
	} else {
		visitNumberValue = len(prevVisits)
	}
	return &VisitorVisits{
		visitNumber: visitNumberValue,
		prevVisits:  prevVisits,
	}
}

func (vv *VisitorVisits) Localize(timeStarted int64) *VisitorVisits {
	var timeSincePreviousVisit int64
	for _, visit := range vv.prevVisits {
		timeDelta := timeStarted - visit.timeLastActivity
		if timeDelta >= 0 {
			timeSincePreviousVisit = timeDelta
			break
		}
	}
	return &VisitorVisits{
		visitNumber:            vv.visitNumber,
		prevVisits:             vv.prevVisits,
		timeStarted:            timeStarted,
		timeSincePreviousVisit: timeSincePreviousVisit,
	}
}

func (vv *VisitorVisits) VisitNumber() int {
	return vv.visitNumber
}

func (vv *VisitorVisits) PrevVisits() []Visit {
	if vv != nil {
		return vv.prevVisits
	}
	return nil
}

func (vv *VisitorVisits) TimeStarted() int64 {
	return vv.timeStarted
}

func (vv *VisitorVisits) TimeSincePreviousVisit() int64 {
	return vv.timeSincePreviousVisit
}

func (vv *VisitorVisits) QueryEncode() string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPEventType, visitorVisitsEventType)
	qb.Append(utils.QPVisitNumber, strconv.Itoa(vv.visitNumber))
	qb.Append(utils.QPTimeSincePreviousVisit, strconv.FormatInt(vv.timeSincePreviousVisit, 10))
	qb.Append(utils.QPNonce, vv.Nonce())
	return qb.String()
}

func (vv VisitorVisits) String() string {
	return fmt.Sprintf(
		"VisitorVisits{visitNumber:%v,prevVisits:%v,timeStarted:%v,timeSincePreviousVisit:%v}",
		vv.visitNumber, vv.prevVisits, vv.timeStarted, vv.timeSincePreviousVisit,
	)
}

func (vv *VisitorVisits) DataType() DataType {
	return DataTypeVisitorVisits
}

type Visit struct {
	timeStarted      int64
	timeLastActivity int64
}

func NewVisit(timeStarted int64, timeLastActivity ...int64) Visit {
	timeLastActivityValue := timeStarted
	if len(timeLastActivity) > 0 {
		timeLastActivityValue = timeLastActivity[0]
	}
	return Visit{
		timeStarted:      timeStarted,
		timeLastActivity: timeLastActivityValue,
	}
}

func (v Visit) TimeStarted() int64 {
	return v.timeStarted
}

func (v Visit) TimeLastActivity() int64 {
	return v.timeLastActivity
}

func (v Visit) String() string {
	return fmt.Sprintf(
		"Visit{timeStarted:%v,timeLastActivity:%v,}",
		v.timeStarted, v.timeLastActivity,
	)
}
