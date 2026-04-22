package types

import (
	"fmt"

	"github.com/Kameleoon/client-go/v3/utils"
)

type Variation struct {
	Key          string
	Name         string
	VariationID  *int
	ExperimentID *int
	Variables    map[string]Variable
}

// Intended for internal use only.
func (Variation) BuildFromInternal(
	sourceVariation *VariationFeatureFlag,
	variationID *int,
	experimentID *int,
) Variation {
	variables := map[string]Variable{}
	key := ""
	name := ""

	if sourceVariation != nil {
		key = sourceVariation.Key
		name = sourceVariation.Name
		variables = make(map[string]Variable, len(sourceVariation.Variables))
		for _, internalVariable := range sourceVariation.Variables {
			variables[internalVariable.Key] = Variable{
				Key:   internalVariable.Key,
				Type:  internalVariable.Type,
				Value: internalVariable.GetVariableValue(),
			}
		}
	}

	return Variation{
		Key:          key,
		Name:         name,
		VariationID:  utils.Reref(variationID),
		ExperimentID: utils.Reref(experimentID),
		Variables:    variables,
	}
}

func (v Variation) IsActive() bool {
	return v.Key != string(VariationOff)
}

func (v Variation) String() string {
	return fmt.Sprintf(
		"Variation{Key:'%v',Name:'%v',VariationID:%v,ExperimentID:%v,Variables:%v}",
		v.Key, v.Name, v.VariationID, v.ExperimentID, v.Variables,
	)
}
