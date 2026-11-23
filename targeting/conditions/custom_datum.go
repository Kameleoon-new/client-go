package conditions

import (
	"github.com/Kameleoon/client-go/v3/logging"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
	"github.com/segmentio/encoding/json"
)

func NewCustomDatum(c types.TargetingCondition) *CustomDatum {
	if c.Value == nil {
		c.Value = ""
	}
	return &CustomDatum{
		TargetingConditionBase: c.TargetingConditionBase,
		Operator:               c.Operator,
		Value:                  c.Value,
		Index:                  c.Index,
		Include:                c.Include,
	}
}

type CustomDatum struct {
	types.TargetingConditionBase
	Value    interface{}         `json:"value"`
	Type     types.TargetingType `json:"targetingType"`
	Operator types.OperatorType  `json:"valueMatchType"`
	Index    string              `json:"customDataIndex"`
	Include  bool                `json:"include"`
	// Prepared from Value on the first check. A condition is shared between goroutines, so sync.Once guards it
	prepareOnce sync.Once
	pattern     *regexp.Regexp
	amongValues []string
}

func (c *CustomDatum) CheckTargeting(targetData interface{}) bool {
	customDataStorage, ok := targetData.(storage.DataMapStorage[int, types.ICustomData])
	if !ok || (customDataStorage == nil) {
		return false
	}
	if id, err := strconv.Atoi(c.Index); err == nil {
		if customData := customDataStorage.Get(id); customData != nil {
			return c.checkTargeting(customData.Values())
		}
	}
	return c.Operator == types.OperatorUndefined
}

func (c *CustomDatum) checkTargeting(customDataValues []string) bool {
	if isComparison(c.Operator) {
		return c.compareNumberValues(customDataValues)
	}
	switch c.Operator {
	case types.OperatorContains:
		value, ok := c.Value.(string)
		if !ok {
			return false
		}
		return c.contains(customDataValues, func(customDataValue string) bool {
			return strings.Contains(customDataValue, value)
		})
	case types.OperatorExact:
		return c.contains(customDataValues, func(customDataValue string) bool {
			return customDataValue == c.Value
		})
	case types.OperatorRegExp:
		c.prepareOnce.Do(c.prepare)
		if c.pattern == nil {
			return false
		}
		return c.contains(customDataValues, c.pattern.MatchString)
	case types.OperatorIsTrue:
		return c.contains(customDataValues, func(customDataValue string) bool {
			val, err := strconv.ParseBool(customDataValue)
			return err == nil && val
		})
	case types.OperatorIsFalse:
		return c.contains(customDataValues, func(customDataValue string) bool {
			val, err := strconv.ParseBool(customDataValue)
			return err == nil && !val
		})
	case types.OperatorIsAmongValues:
		c.prepareOnce.Do(c.prepare)
		return c.contains(customDataValues, func(customDataValue string) bool {
			return utils.SliceContains(c.amongValues, customDataValue)
		})
	case types.OperatorUndefined:
		// The custom data is defined, so the condition is not met
	default:
		logging.Error("Unexpected comparing operation for %s condition: %s", types.TargetingCustomDatum, c.Operator)
	}
	return false
}

func (c *CustomDatum) prepare() {
	switch c.Operator {
	case types.OperatorRegExp:
		c.pattern = compilePattern(c.Value)
	case types.OperatorIsAmongValues:
		c.amongValues = parseAmongValues(c.Value)
	}
}

func compilePattern(value interface{}) *regexp.Regexp {
	valueString, ok := value.(string)
	if !ok {
		return nil
	}
	pattern, err := regexp.Compile(valueString)
	if err != nil {
		logging.Error("Failed to compile pattern %s for 'Custom' condition (value): %s", valueString, err)
		return nil
	}
	return pattern
}

func parseAmongValues(value interface{}) []string {
	var values []interface{}
	valueString, _ := value.(string)
	if err := json.Unmarshal([]byte(valueString), &values); err != nil {
		logging.Error("Failed parse values %v for 'Custom' condition (value): %s", value, err)
		return nil
	}
	amongValues := make([]string, len(values))
	for i, element := range values {
		amongValues[i] = fmt.Sprintf("%v", element)
	}
	return amongValues
}

func (c *CustomDatum) compareNumberValues(customDataValues []string) bool {
	number := c.numberValue()
	// A non-numeric value is NaN, which never matches, so it does not hide the other values
	return c.contains(customDataValues, func(customDataValue string) bool {
		targeted, _ := compare(c.Operator, parseNumber(customDataValue), number)
		return targeted
	})
}

func (c *CustomDatum) numberValue() float64 {
	switch v := c.Value.(type) {
	case string:
		return parseNumber(v)
	case int:
		return float64(v)
	case float64:
		return v
	default:
		return math.NaN()
	}
}

func parseNumber(value string) float64 {
	if number, err := strconv.ParseFloat(value, 64); err == nil {
		return number
	}
	return math.NaN()
}

func (c *CustomDatum) contains(customDataValues []string, callback func(string) bool) bool {
	for _, val := range customDataValues {
		if callback(val) {
			return true
		}
	}
	return false
}

func (c *CustomDatum) String() string {
	return utils.JsonToString(c)
}
