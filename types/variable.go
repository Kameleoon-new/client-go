package types

import (
	"fmt"

	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/segmentio/encoding/json"
)

type Variable struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (v Variable) String() string {
	return fmt.Sprintf(
		"Variable{Key:'%v',Type:'%v',Value:%v}",
		v.Key,
		v.Type,
		logging.ObjectToString(v.Value),
	)
}

func (v *Variable) GetVariableValue() interface{} {
	var value interface{}
	switch v.Type {
	case "JSON":
		if valueString, ok := v.Value.(string); ok {
			if err := json.Unmarshal([]byte(valueString), &value); err != nil {
				value = nil
			}
		}
	default:
		value = v.Value
	}
	return value
}
