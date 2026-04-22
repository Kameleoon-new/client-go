package types

import (
	"fmt"
)

type DataFile struct {
	FeatureFlags map[string]FeatureFlag
	DateModified int64
}

// Intended for internal use only.
func (DataFile) BuildFromInternal(sourceDataFile IDataFile) *DataFile {
	if sourceDataFile == nil {
		return &DataFile{FeatureFlags: map[string]FeatureFlag{}}
	}

	internalFeatureFlags := sourceDataFile.GetFeatureFlags()
	featureFlags := make(map[string]FeatureFlag, len(internalFeatureFlags))
	for featureKey, internalFeatureFlag := range internalFeatureFlags {
		featureFlags[featureKey] = (FeatureFlag{}).BuildFromInternal(internalFeatureFlag)
	}
	return &DataFile{FeatureFlags: featureFlags, DateModified: sourceDataFile.DateModified()}
}

func (df DataFile) String() string {
	return fmt.Sprintf("DataFile{FeatureFlags:%v,DateModified:%v}", df.FeatureFlags, df.DateModified)
}
