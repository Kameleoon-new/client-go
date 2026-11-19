package configuration

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/types"
)

const (
	consentTypeRequired = "REQUIRED"

	configUpdateModePolling   = "POLLING"
	configUpdateModeStreaming = "STREAMING"
	// Streaming for server-side SDKs only, this one included
	configUpdateModeStreamingServerOnly = "STREAMING_SERVER_ONLY"
)

type Settings struct {
	streamingMode                      bool
	isConsentRequired                  bool
	blockingBehaviourIfConsentNotGiven types.ConsentBlockingBehaviour
	dataApiDomain                      string
}

func (s Settings) String() string {
	return fmt.Sprintf("Settings{streamingMode:%v,isConsentRequired:%v,dataApiDomain:'%v'}",
		s.streamingMode, s.isConsentRequired, s.dataApiDomain)
}

func (s *Settings) UnmarshalJSON(data []byte) error {
	var sm = struct {
		ConfigUpdateMode      string `json:"configUpdateMode"`
		ConsentType           string `json:"consentType"`
		ConsentOptOutBehavior string `json:"consentOptOutBehavior"`
		DataApiDomain         string `json:"dataApiDomain,omitempty"`
	}{}
	if err := json.Unmarshal(data, &sm); err != nil {
		return err
	}
	s.streamingMode = isStreamingConfigUpdateMode(sm.ConfigUpdateMode)
	s.isConsentRequired = sm.ConsentType == consentTypeRequired
	s.blockingBehaviourIfConsentNotGiven = types.ConsentBlockingBehaviourFromStr(sm.ConsentOptOutBehavior)
	s.dataApiDomain = sm.DataApiDomain
	return nil
}

func isStreamingConfigUpdateMode(configUpdateMode string) bool {
	switch strings.ToUpper(configUpdateMode) {
	case configUpdateModeStreaming, configUpdateModeStreamingServerOnly:
		return true
	case configUpdateModePolling:
		return false
	default:
		logging.Warning("Unexpected ConfigUpdateMode %s", configUpdateMode)
		return false
	}
}

// StreamingMode is true when configuration updates are to be received as streaming events
// rather than polled.
func (s Settings) StreamingMode() bool {
	return s.streamingMode
}

func (s Settings) IsConsentRequired() bool {
	return s.isConsentRequired
}

func (s Settings) BlockingBehaviourIfConsentNotGiven() types.ConsentBlockingBehaviour {
	return s.blockingBehaviourIfConsentNotGiven
}

func (s Settings) DataApiDomain() string {
	return s.dataApiDomain
}
