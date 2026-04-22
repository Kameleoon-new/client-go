package types

import (
	"fmt"
)

type Rule struct {
	Variations map[string]Variation
}

// Intended for internal use only.
func (Rule) BuildFromInternal(sourceRule IRule, variations map[string]Variation) Rule {
	experiment := sourceRule.GetRuleBase().Experiment
	ruleVars := make(map[string]Variation, len(experiment.VariationsByExposition))
	for _, varByExp := range experiment.VariationsByExposition {
		baseVariation, exists := variations[varByExp.VariationKey]
		if !exists {
			continue
		}
		ruleVars[baseVariation.Key] = Variation{
			Key:          baseVariation.Key,
			Name:         baseVariation.Name,
			VariationID:  varByExp.VariationID,
			ExperimentID: &experiment.ExperimentId,
			Variables:    baseVariation.Variables,
		}
	}
	return Rule{Variations: ruleVars}
}

func (r Rule) String() string {
	return fmt.Sprintf("Rule{Variations:%v}", r.Variations)
}
