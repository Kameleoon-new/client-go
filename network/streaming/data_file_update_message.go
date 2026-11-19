package streaming

import (
	"encoding/json"
	"errors"
)

// DataFileUpdateMessage is a streaming event about the current configuration: the server
// pushes one when a configuration is published and one on every (re)connection.
type DataFileUpdateMessage struct {
	// TimeStamp is the `dateModified` of the configuration the message refers to, in milliseconds.
	TimeStamp int64
}

func (m *DataFileUpdateMessage) UnmarshalJSON(data []byte) error {
	var model struct {
		Ts   *int64 `json:"ts"`
		TsMs *int64 `json:"tsMs"`
	}
	if err := json.Unmarshal(data, &model); err != nil {
		return err
	}
	// Both hold milliseconds; `ts` is only read while the server still sends events without `tsMs`
	switch {
	case model.TsMs != nil:
		m.TimeStamp = *model.TsMs
	case model.Ts != nil:
		m.TimeStamp = *model.Ts
	default:
		return errors.New("timestamp is missing")
	}
	return nil
}
