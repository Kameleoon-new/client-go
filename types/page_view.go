package types

import (
	"fmt"
	"strings"

	"github.com/Kameleoon/client-go/v3/utils"
)

const pageViewEventType = "page"

type PageView struct {
	duplicationSafeSendableBase
	url       string
	title     string
	referrers []int
}

func NewPageView(url string, referrers ...int) *PageView {
	return NewPageViewWithTitle(url, "", referrers...)
}
func NewPageViewWithTitle(url string, title string, referrers ...int) *PageView {
	pv := &PageView{
		url:       url,
		title:     title,
		referrers: referrers,
	}
	pv.initSendable()
	return pv
}

func (pv PageView) String() string {
	return fmt.Sprintf(
		"PageView{url:'%v',title:'%v',referrers:%v}",
		pv.url,
		pv.title,
		pv.referrers,
	)
}

func (pv *PageView) dataRestriction() {
	// This method is required to separate external type `Data` from `BaseData` types
}

func (pv *PageView) URL() string {
	return pv.url
}

func (pv *PageView) Title() string {
	return pv.title
}

func (pv *PageView) Referrers() []int {
	return pv.referrers
}

func (pv *PageView) QueryEncode() string {
	if len(pv.url) == 0 {
		return ""
	}
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPEventType, pageViewEventType)
	qb.Append(utils.QPHref, pv.url)
	qb.Append(utils.QPTitle, pv.title)
	if len(pv.referrers) > 0 {
		qb.Append(utils.QPReferrersIndices, pv.encodeReferrers())
	}
	qb.Append(utils.QPNonce, pv.Nonce())
	return qb.String()
}
func (pv *PageView) encodeReferrers() string {
	return strings.ReplaceAll(fmt.Sprint(pv.referrers), " ", ",")
}

func (pv *PageView) DataType() DataType {
	return DataTypePageView
}
