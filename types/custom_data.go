package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kameleoon/client-go/v3/logging"

	"github.com/Kameleoon/client-go/v3/utils"
)

type ICustomData interface {
	Data
	Sendable
	Index() int
	Name() string
	Values() []string
	Overwrite() bool
}

const customDataEventType = "customData"
const CustomDataUndefinedIndex = -1

type CustomDataOptParams struct {
	overwrite bool
}

func NewCustomDataOptParams() CustomDataOptParams {
	return CustomDataOptParams{overwrite: true}
}
func (p CustomDataOptParams) Overwrite(value bool) CustomDataOptParams {
	p.overwrite = value
	return p
}

type CustomData struct {
	duplicationUnsafeSendableBase
	index     int
	name      string
	values    []string
	overwrite bool
}

func newCustomData(index int, name string, overwrite bool, values []string) *CustomData {
	return &CustomData{
		index:     index,
		name:      name,
		values:    values,
		overwrite: overwrite,
	}
}

func NewCustomData(index int, values ...string) *CustomData {
	return newCustomData(index, "", true, values)
}

func NewCustomDataWithOptParams(index int, params CustomDataOptParams, values ...string) *CustomData {
	return newCustomData(index, "", params.overwrite, values)
}

func NewNamedCustomData(name string, values ...string) *CustomData {
	return newCustomData(CustomDataUndefinedIndex, name, true, values)
}

func NewNamedCustomDataWithOptParams(name string, params CustomDataOptParams, values ...string) *CustomData {
	return newCustomData(CustomDataUndefinedIndex, name, params.overwrite, values)
}

func (cd *CustomData) NamedToIndexed(index int) *CustomData {
	return newCustomData(index, cd.name, cd.overwrite, cd.values)
}

func (cd CustomData) String() string {
	return fmt.Sprintf(
		"CustomData{index:%d,name:'%v',values:%s,overwrite:%v}",
		cd.index, cd.name, logging.ObjectToString(cd.values), cd.overwrite,
	)
}

func (cd *CustomData) dataRestriction() {
	// This method is required to separate external type `Data` from `BaseData` types
}

// Deprecated: Please use `Index` instead
func (cd *CustomData) ID() int {
	return cd.index
}

func (cd *CustomData) Index() int {
	return cd.index
}

func (cd *CustomData) Name() string {
	return cd.name
}

func (cd *CustomData) Values() []string {
	return cd.values
}

func (cd *CustomData) Overwrite() bool {
	return cd.overwrite
}

func (cd *CustomData) QueryEncode() string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPEventType, customDataEventType)
	qb.Append(utils.QPIndex, strconv.Itoa(cd.index))
	qb.Append(utils.QPValuesCountMap, cd.encodeValues())
	qb.Append(utils.QPOverwrite, strconv.FormatBool(cd.overwrite))
	qb.Append(utils.QPNonce, cd.Nonce())
	return qb.String()
}

func (cd *CustomData) encodeValues() string {
	escaping := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
	)
	sb := strings.Builder{}
	sb.WriteString("{")
	for i, value := range cd.values {
		if i > 0 {
			sb.WriteString(",\"")
		} else {
			sb.WriteString("\"")
		}
		sb.WriteString(escaping.Replace(value))
		sb.WriteString("\":1")
	}
	sb.WriteRune('}')
	return sb.String()
}

func (cd *CustomData) DataType() DataType {
	return DataTypeCustom
}
