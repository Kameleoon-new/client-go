package tracking

import (
	"fmt"

	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/Kameleoon/client-go/v3/utils"
)

type TrackingBuilder struct {
	built bool

	visitorCodes     VisitorCodeCollection
	dataFile         types.IDataFile
	visitorManager   storage.VisitorManager
	requestSizeLimit int
	totalSize        int

	// Result
	visitorCodesToSend []string
	visitorCodesToKeep []string
	trackingLines      []string
	unsentVisitorData  []types.Sendable
}

func NewTrackingBuilder(
	visitorCodes VisitorCodeCollection, dataFile types.IDataFile, visitorManager storage.VisitorManager,
	requestSizeLimit int,
) *TrackingBuilder {
	return &TrackingBuilder{
		visitorCodes:     visitorCodes,
		dataFile:         dataFile,
		visitorManager:   visitorManager,
		requestSizeLimit: requestSizeLimit,
	}
}

func (tb *TrackingBuilder) VisitorCodesToSend() []string {
	return tb.visitorCodesToSend
}
func (tb *TrackingBuilder) VisitorCodesToKeep() []string {
	return tb.visitorCodesToKeep
}
func (tb *TrackingBuilder) TrackingLines() []string {
	return tb.trackingLines
}
func (tb *TrackingBuilder) UnsentVisitorData() []types.Sendable {
	return tb.unsentVisitorData
}

func logVisitorTrackSending(visitorCode string, isConsentGiven bool, data []types.Sendable) {
	logging.Debug(
		"Sending tracking request for unsent data %v of visitor %s with given (or not required) consent %s",
		data, visitorCode, isConsentGiven)
}
func logVisitorTrackNoData(visitorCode string, isConsentGiven bool) {
	logging.Debug("No data to send for visitor %s with given (or not required) consent %s",
		visitorCode, isConsentGiven)
}

// Not thread-safe
func (tb *TrackingBuilder) Build() {
	if tb.built {
		return
	}
	tb.visitorCodes.Range(func(visitorCode string) bool {
		if tb.totalSize <= tb.requestSizeLimit {
			visitor := tb.visitorManager.PeekVisitor(visitorCode)
			isConsentGiven := tb.isConsentGiven(visitor)
			data := tb.collectTrackingData(visitorCode, visitor, isConsentGiven)
			if len(data) > 0 {
				logVisitorTrackSending(visitorCode, isConsentGiven, data)
				tb.visitorCodesToSend = append(tb.visitorCodesToSend, visitorCode)
				tb.unsentVisitorData = append(tb.unsentVisitorData, data...)
			} else {
				logVisitorTrackNoData(visitorCode, isConsentGiven)
			}
		} else {
			tb.visitorCodesToKeep = append(tb.visitorCodesToKeep, visitorCode)
		}
		return true
	})
	tb.built = true
}

func (tb *TrackingBuilder) isConsentGiven(visitor storage.Visitor) bool {
	return !tb.dataFile.Settings().IsConsentRequired() ||
		((visitor != nil) && (visitor.LegalConsent() == types.LegalConsentGiven))
}

func (tb *TrackingBuilder) collectTrackingData(
	visitorCode string, visitor storage.Visitor, isConsentGiven bool,
) []types.Sendable {
	var useMappingValue bool
	useMappingValue, visitor = tb.createSelfVisitorLinkIfRequired(visitorCode, visitor)
	logging.Info(func() string {
		idType := "visitor code"
		if useMappingValue {
			idType = "mapping value"
		}
		return fmt.Sprintf("'%s' was used as a %s for visitor data tracking.\n", visitorCode, idType)
	})
	unsentData := tb.getUnsentVisitorData(visitor, isConsentGiven)
	tb.collectTrackingLines(unsentData, visitorCode, visitor, useMappingValue)
	return unsentData
}

func (tb *TrackingBuilder) createSelfVisitorLinkIfRequired(
	visitorCode string, visitor storage.Visitor,
) (bool, storage.Visitor) {
	isMapped := (visitor != nil) && (visitor.MappingIdentifier() != nil)
	isUniqueIdentifier := (visitor != nil) && visitor.IsUniqueIdentifier()
	// need to find if anonymous visitor is behind unique (anonym doesn't exist if MappingIdentifier == nil)
	if isUniqueIdentifier && !isMapped {
		// We haven't anonymous behind, in this case we should create "fake" anonymous with id == visitorCode
		// and link it with with mapping value == visitorCode (like we do as we have real anonymous visitor)
		visitor = tb.visitorManager.AddData(
			visitorCode, types.NewCustomData(tb.dataFile.CustomDataInfo().MappingIdentifierIndex(), visitorCode),
		)
	}
	var useMappingValue bool
	if isUniqueIdentifier {
		mappingIdentifier := visitor.MappingIdentifier()
		useMappingValue = (mappingIdentifier != nil) && (visitorCode != *mappingIdentifier)
	}
	return useMappingValue, visitor
}

func (tb *TrackingBuilder) getUnsentVisitorData(visitor storage.Visitor, isConsentGiven bool) []types.Sendable {
	var unsentData []types.Sendable
	if visitor != nil {
		if isConsentGiven {
			visitor.EnumerateSendableData(func(s types.Sendable) bool {
				if s.Unsent() {
					unsentData = append(unsentData, s)
				}
				return true
			})
		} else {
			visitor.Variations().Enumerate(func(av *types.AssignedVariation) bool {
				if av.Unsent() && (av.RuleType() == types.RuleTypeTargetedDelivery) {
					unsentData = append(unsentData, av)
				}
				return true
			})
			visitor.Conversions().Enumerate(func(c *types.Conversion) bool {
				if c.Unsent() {
					unsentData = append(unsentData, c)
				}
				return true
			})
		}
	}
	if (len(unsentData) == 0) && isConsentGiven {
		unsentData = append(unsentData, types.NewActivityEvent())
	}
	return unsentData
}

func (tb *TrackingBuilder) collectTrackingLines(
	unsentVisitorData []types.Sendable, visitorCode string, visitor storage.Visitor, useMappingValue bool,
) {
	visitorCodeParam := makeVisitorCodeParam(visitorCode, useMappingValue)
	var userAgent string
	if visitor != nil {
		userAgent = visitor.UserAgent()
	}
	for _, s := range unsentVisitorData {
		line := s.QueryEncode()
		if line != "" {
			line = addLineParams(line, visitorCodeParam, userAgent)
			tb.trackingLines = append(tb.trackingLines, line)
			tb.totalSize += len(line)
			userAgent = ""
		}
	}
}
func makeVisitorCodeParam(visitorCode string, useMappingValue bool) string {
	qb := utils.NewQueryBuilder()
	if useMappingValue {
		qb.Append(utils.QPMappingValue, visitorCode)
	} else {
		qb.Append(utils.QPVisitorCode, visitorCode)
	}
	return qb.String()
}
func addLineParams(trackingLine string, visitorCodeParam string, userAgent string) string {
	trackingLine += "&" + visitorCodeParam
	if userAgent != "" {
		userAgentParam := utils.NewQueryBuilder().Append(utils.QPUserAgent, userAgent).String()
		trackingLine += "&" + userAgentParam
	}
	return trackingLine
}
