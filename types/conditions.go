package types

import (
	"math"
	"strings"

	"github.com/segmentio/encoding/json"
)

type Condition interface {
	GetType() TargetingType
	GetInclude() bool
	CheckTargeting(interface{}) bool
	String() string
}

type ConditionsData struct {
	FirstLevelOrOperators []bool                 `json:"firstLevelOrOperators"`
	FirstLevel            []ConditionsFirstLevel `json:"firstLevel"`
}

type ConditionsFirstLevel struct {
	OrOperators []bool               `json:"orOperators"`
	Conditions  []TargetingCondition `json:"conditions"`
}

type BrowserConditionType string

const (
	BrowserConditionTypeChrome  BrowserConditionType = "CHROME"
	BrowserConditionTypeIE      BrowserConditionType = "IE"
	BrowserConditionTypeFirefox BrowserConditionType = "FIREFOX"
	BrowserConditionTypeSafari  BrowserConditionType = "SAFARI"
	BrowserConditionTypeOpera   BrowserConditionType = "OPERA"
	BrowserConditionTypeOther   BrowserConditionType = "OTHER"
)

type TargetingConditionBase struct {
	Type    TargetingType `json:"targetingType"`
	Include bool          `json:"isInclude,omitempty"`
}

type targetingCondition struct {
	TargetingConditionBase
	Value              interface{}          `json:"value,omitempty"`
	Operator           OperatorType         `json:"valueMatchType,omitempty"`
	Index              string               `json:"customDataIndex,omitempty"`
	Experiment         int                  `json:"experiment,omitempty"`
	Variation          int                  `json:"variation,omitempty"`
	VariationMatchType OperatorType         `json:"variationMatchType,omitempty"`
	VersionMatchType   OperatorType         `json:"versionMatchType,omitempty"`
	Browser            BrowserConditionType `json:"browser,omitempty"`
	Version            string               `json:"version,omitempty"`
	Device             DeviceType           `json:"device,omitempty"`
	VisitorCode        string               `json:"visitorCode,omitempty"`
	MatchType          OperatorType         `json:"matchType,omitempty"`
	SdkLanguage        string               `json:"sdkLanguage,omitempty"`
	Title              string               `json:"title,omitempty"`
	Url                string               `json:"url,omitempty"`
	GoalId             int                  `json:"goalId,omitempty"`
	Name               string               `json:"name,omitempty"`
	NameMatchType      OperatorType         `json:"nameMatchType,omitempty"`
	Country            string               `json:"country,omitempty"`
	Region             string               `json:"region,omitempty"`
	City               string               `json:"city,omitempty"`
	OS                 string               `json:"os,omitempty"`
	SegmentId          int                  `json:"segmentId,omitempty"`
	FeatureFlagId      int                  `json:"featureFlagId,omitempty"`
	VariationKey       string               `json:"variationKey,omitempty"`
	RuleId             int                  `json:"ruleId,omitempty"`
	PageCount          int                  `json:"pageCount,omitempty"`
	CountInMillis      int64                `json:"countInMillis,omitempty"`
	VisitCount         int                  `json:"visitCount,omitempty"`
	VisitScope         string               `json:"visitScope,omitempty"`
	VisitorType        string               `json:"visitorType,omitempty"`
	KeyMomentId        int                  `json:"keyMomentId,omitempty"`
	LowerBound         float64              `json:"lowerBound,omitempty"`
	UpperBound         float64              `json:"upperBound,omitempty"`
	PersonalizationId  int                  `json:"personalizationId,omitempty"`
	CampaignType       string               `json:"campaignType,omitempty"`
	VariationId        int                  `json:"variationId,omitempty"`
	ExperimentId       int                  `json:"experimentId,omitempty"`
}

type TargetingCondition targetingCondition

const (
	UndefinedCountInMillisValue = -1
	UndefinedVisitCountValue    = -1
	UndefinedGoalId             = -1
	UndefinedPersonalizationId  = -1
	UndefinedVariationId        = -1
	UndefinedExperimentId       = -1
)

func (c *TargetingCondition) UnmarshalJSON(data []byte) error {
	c.Include = true
	c.PageCount = math.MaxInt
	c.CountInMillis = UndefinedCountInMillisValue
	c.VisitCount = UndefinedVisitCountValue
	c.GoalId = UndefinedGoalId
	c.PersonalizationId = UndefinedPersonalizationId
	c.VariationId = UndefinedVariationId
	c.ExperimentId = UndefinedExperimentId
	return json.Unmarshal(data, (*targetingCondition)(c))
}

func (c TargetingConditionBase) String() string {
	var s strings.Builder
	b, _ := json.Marshal(c)
	s.Write(b)
	return s.String()
}

func (c TargetingConditionBase) GetType() TargetingType {
	return c.Type
}

func (c TargetingConditionBase) GetInclude() bool {
	return c.Include
}

func (c TargetingConditionBase) CheckTargeting(_ []Data) bool {
	return true
}
