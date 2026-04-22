package configuration

import (
	"fmt"

	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/types"
)

type Configuration struct {
	CustomDataInfo *types.CustomDataInfo `json:"customData"`
	Holdout        *types.Experiment     `json:"holdout"`
	Settings       Settings              `json:"configuration"`
	FeatureFlags   []FeatureFlag         `json:"featureFlags"`
	Segments       []types.SegmentBase   `json:"segments"`
	DateModified   int64                 `json:"dateModified"`
}

func (c Configuration) String() string {
	return fmt.Sprintf(
		"Configuration{CustomDataInfo:%v,Settings:%v,FeatureFlags:%v,Segments:%v,DateModified:%v}",
		c.CustomDataInfo, c.Settings, logging.ObjectToString(c.FeatureFlags), logging.ObjectToString(c.Segments),
		c.DateModified,
	)
}
