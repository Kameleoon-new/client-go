package types

type IDataFile interface {
	LastModified() string
	DateModified() int64
	CustomDataInfo() *CustomDataInfo
	Holdout() *Experiment
	Settings() Settings
	Segments() map[int]Segment
	AudienceTrackingSegments() []Segment
	GetFeatureFlags() map[string]IFeatureFlag
	GetOrderedFeatureFlags() []IFeatureFlag
	GetFeatureFlag(featureKey string) (IFeatureFlag, error)
	EnsureEnvironmentEnabled(featureFlag IFeatureFlag) error
	MEGroups() map[string]MEGroup

	HasAnyTargetedDeliveryRule() bool
	GetFeatureFlagById(featureFlagId int) IFeatureFlag
	GetRuleInfoByExpId(experimentId int) (RuleInfo, bool)
	GetVariation(variationId int) *VariationByExposition
	HasExperimentJsCssVariable(experimentId int) bool
}

type RuleInfo struct {
	FeatureFlag IFeatureFlag
	Rule        IRule
}
