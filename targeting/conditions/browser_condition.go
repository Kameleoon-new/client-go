package conditions

import (
	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
)

func NewBrowserCondition(c types.TargetingCondition) *BrowserCondition {
	return &BrowserCondition{
		TargetingConditionBase: types.TargetingConditionBase{
			Type:    c.Type,
			Include: c.Include,
		},
		Browser:          c.Browser,
		Version:          c.Version,
		VersionMatchType: c.VersionMatchType,
	}
}

type BrowserCondition struct {
	types.TargetingConditionBase
	Browser          types.BrowserConditionType `json:"browser"`
	Version          string                     `json:"version,omitempty"`
	VersionMatchType types.OperatorType         `json:"versionMatchType,omitempty"`
}

func (c *BrowserCondition) CheckTargeting(targetData interface{}) bool {
	browser, ok := targetData.(*types.Browser)
	return ok && (browser != nil) && c.checkTargeting(browser)
}

func (c *BrowserCondition) checkTargeting(browser *types.Browser) bool {
	// return false, if browser types are not equal
	if c.browserStringToInt() != browser.Type() {
		return false
	}
	// return true, browser types are equal and version isn't defined
	if len(c.Version) == 0 {
		return true
	}
	// check the version because it's defined in condition
	version, err := utils.NewVersionFromString(c.Version)
	if err != nil {
		logging.Error("Failed to parse version %s for 'Browser' condition: %s", c.Version, err)
		return false
	}
	floatVersion, err := version.ToFloat()
	if err != nil {
		return false
	}

	targeted, isComparison := compare(c.VersionMatchType, browser.Version(), floatVersion)
	if !isComparison {
		logging.Error("Unexpected comparing operation for 'Browser' condition: %s", c.VersionMatchType)
	}
	return targeted
}

func (c BrowserCondition) String() string {
	return utils.JsonToString(c)
}

func (c *BrowserCondition) browserStringToInt() types.BrowserType {
	switch c.Browser {
	case types.BrowserConditionTypeChrome:
		return types.BrowserTypeChrome
	case types.BrowserConditionTypeIE:
		return types.BrowserTypeIE
	case types.BrowserConditionTypeFirefox:
		return types.BrowserTypeFirefox
	case types.BrowserConditionTypeSafari:
		return types.BrowserTypeSafari
	case types.BrowserConditionTypeOpera:
		return types.BrowserTypeOpera
	}
	return types.BrowserTypeOther
}
