package types

import (
	"fmt"
)

type FeatureFlag struct {
	Variations           map[string]Variation
	IsEnvironmentEnabled bool
	Rules                []Rule
	DefaultVariationKey  string
}

// Intended for internal use only.
func (FeatureFlag) BuildFromInternal(sourceFeatureFlag IFeatureFlag) FeatureFlag {
	internalVariations := sourceFeatureFlag.GetVariations()
	variations := make(map[string]Variation, len(internalVariations))
	for _, internalVariation := range internalVariations {
		variations[internalVariation.Key] = (Variation{}).BuildFromInternal(&internalVariation, nil, nil)
	}

	internalRules := sourceFeatureFlag.GetRules()
	rules := make([]Rule, len(internalRules))
	for ruleIndex, internalRule := range internalRules {
		rules[ruleIndex] = (Rule{}).BuildFromInternal(internalRule, variations)
	}

	return FeatureFlag{
		Variations:           variations,
		IsEnvironmentEnabled: sourceFeatureFlag.GetEnvironmentEnabled(),
		Rules:                rules,
		DefaultVariationKey:  sourceFeatureFlag.GetDefaultVariationKey(),
	}
}

func (ff FeatureFlag) DefaultVariation() Variation {
	return ff.Variations[ff.DefaultVariationKey]
}

func (ff FeatureFlag) String() string {
	return fmt.Sprintf(
		"FeatureFlag{Variations:%v,IsEnvironmentEnabled:%v,Rules:%v,DefaultVariationKey:'%v'}",
		ff.Variations, ff.IsEnvironmentEnabled, ff.Rules, ff.DefaultVariationKey,
	)
}
