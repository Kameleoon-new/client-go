package conditions

import (
	"strings"

	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
)

const (
	minVisitorVisitCount = 2
	maxVisitorVisitCount = 25
)

type VisitScope string

const (
	VisitScopeCurrentVisit VisitScope = "CURRENT_VISIT"
	VisitScopeVisitor      VisitScope = "VISITOR"
)

type visitorScopeCondition struct {
	visitScope VisitScope
	visitCount int
}

func newVisitorScopeCondition(c types.TargetingCondition, defaultVisitScope VisitScope) visitorScopeCondition {
	visitCount := c.VisitCount
	if visitCount <= 0 {
		visitCount = maxVisitorVisitCount
	}
	return visitorScopeCondition{
		visitScope: parseVisitScope(c.VisitScope, defaultVisitScope),
		visitCount: visitCount,
	}
}

func parseVisitScope(value string, defaultVisitScope VisitScope) VisitScope {
	switch strings.ToUpper(value) {
	case string(VisitScopeCurrentVisit):
		return VisitScopeCurrentVisit
	case string(VisitScopeVisitor):
		return VisitScopeVisitor
	default:
		return defaultVisitScope
	}
}

func (c visitorScopeCondition) assignmentThresholdMillis(visitorVisits *types.VisitorVisits) int64 {
	if visitorVisits == nil {
		return 0
	}
	prevVisits := visitorVisits.PrevVisits()
	if c.visitScope == VisitScopeCurrentVisit || c.visitCount < minVisitorVisitCount || len(prevVisits) == 0 {
		return visitorVisits.TimeStarted()
	}
	visitIndex := utils.Clamp(c.visitCount-minVisitorVisitCount, 0, len(prevVisits)-1)
	return prevVisits[visitIndex].TimeStarted()
}
