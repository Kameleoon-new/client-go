package targeting

import (
	"strconv"
	"strings"

	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/targeting/conditions"
	"github.com/Kameleoon/client-go/v3/types"
)

type Tree struct {
	LeftTree   *Tree
	RightTree  *Tree
	Condition  types.Condition
	OrOperator bool
}

func (t *Tree) StringPadding(pads int) string {
	if t == nil {
		return ""
	}
	padding := strings.Repeat("    ", pads)
	var s strings.Builder
	s.WriteString(padding)
	s.WriteString("or_operator: ")
	s.WriteString(strconv.FormatBool(t.OrOperator))
	if t.Condition != nil {
		s.WriteByte('\n')
		s.WriteString(padding)
		s.WriteString("condition: ")
		s.WriteString(t.Condition.String())
	}
	if leftTree := t.LeftTree.StringPadding(pads + 1); len(leftTree) > 0 {
		s.WriteByte('\n')
		s.WriteString(padding)
		s.WriteString("left child:\n")
		s.WriteString(leftTree)
	}
	if rightTree := t.RightTree.StringPadding(pads + 1); len(rightTree) > 0 {
		s.WriteByte('\n')
		s.WriteString(padding)
		s.WriteString("right child:\n")
		s.WriteString(rightTree)
	}
	return s.String()
}

func (t Tree) String() string {
	return t.StringPadding(0)
}

func (t *Tree) CheckTargeting(data types.TargetingDataGetter) bool {
	if t.Condition != nil {
		return t.checkCondition(data)
	}

	var leftTargeted, rightTargeted bool
	if t.LeftTree == nil {
		leftTargeted = true
	} else {
		leftTargeted = t.LeftTree.CheckTargeting(data)
	}
	if t.OrOperator && leftTargeted {
		return leftTargeted
	}
	if t.RightTree == nil {
		rightTargeted = true
	} else {
		rightTargeted = t.RightTree.CheckTargeting(data)
	}
	if t.OrOperator && rightTargeted {
		return rightTargeted
	}
	return leftTargeted && rightTargeted
}

func (t *Tree) checkCondition(data types.TargetingDataGetter) bool {
	td := data(t.Condition.GetType())
	targeted := t.Condition.CheckTargeting(td)
	if !t.Condition.GetInclude() {
		targeted = !targeted
	}
	return targeted
}

func NewTree(cd *types.ConditionsData) *Tree {
	return createFirstLevel(cd)
}

func createFirstLevel(cd *types.ConditionsData) *Tree {
	if len(cd.FirstLevel) == 0 {
		return nil
	}

	var leftTree *Tree
	var leftFirstLevel types.ConditionsFirstLevel
	leftFirstLevel, cd.FirstLevel = cd.FirstLevel[0], cd.FirstLevel[1:]
	leftTree = createSecondLevel(&leftFirstLevel)

	if len(cd.FirstLevel) == 0 {
		return leftTree
	}
	var orOperator bool
	orOperator, cd.FirstLevelOrOperators = cd.FirstLevelOrOperators[0], cd.FirstLevelOrOperators[1:]
	if orOperator {
		return &Tree{
			LeftTree:   leftTree,
			RightTree:  createFirstLevel(cd),
			OrOperator: orOperator,
		}
	}
	var rightFirstLevel types.ConditionsFirstLevel
	rightFirstLevel, cd.FirstLevel = cd.FirstLevel[0], cd.FirstLevel[1:]
	rightTree := createSecondLevel(&rightFirstLevel)
	t := &Tree{
		LeftTree:  leftTree,
		RightTree: rightTree,
	}
	if len(cd.FirstLevel) == 0 {
		return t
	}
	orOperator, cd.FirstLevelOrOperators = cd.FirstLevelOrOperators[0], cd.FirstLevelOrOperators[1:]
	return &Tree{
		LeftTree:   t,
		RightTree:  createFirstLevel(cd),
		OrOperator: orOperator,
	}
}

func createSecondLevel(fl *types.ConditionsFirstLevel) *Tree {
	if len(fl.Conditions) == 0 {
		return nil
	}
	var condition types.TargetingCondition
	condition, fl.Conditions = fl.Conditions[0], fl.Conditions[1:]
	leftTree := &Tree{
		Condition: getCondition(condition),
	}
	if len(fl.Conditions) == 0 {
		return leftTree
	}
	var orOperator bool
	orOperator, fl.OrOperators = fl.OrOperators[0], fl.OrOperators[1:]
	if orOperator {
		return &Tree{
			LeftTree:   leftTree,
			RightTree:  createSecondLevel(fl),
			OrOperator: orOperator,
		}
	}
	condition, fl.Conditions = fl.Conditions[0], fl.Conditions[1:]
	rightTree := &Tree{
		Condition: getCondition(condition),
	}
	t := &Tree{
		LeftTree:  leftTree,
		RightTree: rightTree,
	}
	if len(fl.Conditions) == 0 {
		return t
	}
	orOperator, fl.OrOperators = fl.OrOperators[0], fl.OrOperators[1:]
	return &Tree{
		LeftTree:   t,
		RightTree:  createSecondLevel(fl),
		OrOperator: orOperator,
	}
}

func getCondition(c types.TargetingCondition) types.Condition {
	switch c.GetType() {
	case types.TargetingCustomDatum:
		return conditions.NewCustomDatum(c)
	case types.TargetingBrowser:
		return conditions.NewBrowserCondition(c)
	case types.TargetingDeviceType:
		return conditions.NewDeviceCondition(c)
	case types.TargetingVisitorCode:
		return conditions.NewVisitorCodeCondition(c)
	case types.TargetingSDKLanguage:
		return conditions.NewSdkLanguageCondition(c)
	case types.TargetingApplicationVersion:
		return conditions.NewVersionCondition(c)
	case types.TargetingPageTitle:
		return conditions.NewPageTitleCondition(c)
	case types.TargetingPageUrl:
		return conditions.NewPageUrlCondition(c)
	case types.TargetingPageViews:
		return conditions.NewPageViewNumberCondition(c)
	case types.TargetingPreviousPage:
		return conditions.NewPreviousPageCondition(c)
	case types.TargetingConversions:
		return conditions.NewConversionCondition(c)
	case types.TargetingTargetFeatureFlag:
		return conditions.NewTargetFeatureFlagCondition(c)
	case types.TargetingTargetExperiment:
		return conditions.NewTargetExperimentCondition(c)
	case types.TargetingTargetPersonalization:
		return conditions.NewTargetPersonalizationCondition(c)
	case types.TargetingExclusiveExperiment:
		return conditions.NewExclusiveExperimentCondition(c)
	case types.TargetingCookie:
		return conditions.NewCookieCondition(c)
	case types.TargetingGeolocation:
		return conditions.NewGeolocationCondition(c)
	case types.TargetingOperatingSystem:
		return conditions.NewOperatingSystemCondition(c)
	case types.TargetingSegment:
		return conditions.NewSegmentCondition(c)
	case types.TargetingVisits:
		return conditions.NewVisitNumberTotalCondition(c)
	case types.TargetingSameDayVisits:
		return conditions.NewVisitNumberTodayCondition(c)
	case types.TargetingNewVisitors:
		return conditions.NewVisitorNewReturnCondition(c)
	case types.TargetingFirstVisit:
		fallthrough
	case types.TargetingLastVisit:
		return conditions.NewTimeElapsedSinceVisitCondition(c)
	case types.TargetingHeatSlice:
		return conditions.NewKcsHeatRangeCondition(c)
	}
	logging.Info("Unexpected targeting condition type '%s'", c.GetType())
	return conditions.NewUnknownCondition(c)
}
