package types

type Settings interface {
	StreamingMode() bool
	IsConsentRequired() bool
	BlockingBehaviourIfConsentNotGiven() ConsentBlockingBehaviour
	DataApiDomain() string
}
