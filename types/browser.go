package types

import (
	"fmt"

	"github.com/Kameleoon/client-go/v3/utils"
)

type BrowserType int

const (
	BrowserTypeChrome  BrowserType = 0
	BrowserTypeIE      BrowserType = 1
	BrowserTypeFirefox BrowserType = 2
	BrowserTypeSafari  BrowserType = 3
	BrowserTypeOpera   BrowserType = 4
	BrowserTypeOther   BrowserType = 5
)

func ParseBrowserType(s string) (BrowserType, bool) {
	switch s {
	case "CHROME":
		return BrowserTypeChrome, true
	case "INTERNET_EXPLORER":
		return BrowserTypeIE, true
	case "FIREFOX":
		return BrowserTypeFirefox, true
	case "SAFARI":
		return BrowserTypeSafari, true
	case "OPERA":
		return BrowserTypeOpera, true
	case "OTHER":
		return BrowserTypeOther, true
	}
	return -1, false
}

const browserEventType = "staticData"

type Browser struct {
	duplicationUnsafeSendableBase
	browserType BrowserType
	version     float32
}

func NewBrowser(browserType BrowserType, version ...float32) *Browser {
	var versionValue float32
	if len(version) > 0 {
		versionValue = version[0]
	}
	return &Browser{browserType: browserType, version: versionValue}
}

func (b Browser) String() string {
	return fmt.Sprintf("Browser{browserType:%v,version:%v}", b.browserType, b.version)
}

func (b *Browser) dataRestriction() {
	// This method is required to separate external type `Data` from `BaseData` types
}

func (b *Browser) Type() BrowserType {
	return b.browserType
}

func (b *Browser) Version() float32 {
	return b.version
}

func (b *Browser) QueryEncode() string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPEventType, browserEventType)
	qb.Append(utils.QPBrowserIndex, utils.WritePositiveInt(int(b.browserType)))
	if b.version != 0 {
		qb.Append(utils.QPBrowserVersion, fmt.Sprintf("%f", b.version))
	}
	qb.Append(utils.QPNonce, b.Nonce())
	return qb.String()
}

func (b *Browser) DataType() DataType {
	return DataTypeBrowser
}
