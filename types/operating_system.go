package types

import (
	"fmt"

	"github.com/Kameleoon/client-go/v3/utils"
)

const operatingSystemEventType = "staticData"

type OperatingSystemType byte

const (
	OperatingSystemTypeWindows      OperatingSystemType = 0
	OperatingSystemTypeMac          OperatingSystemType = 1
	OperatingSystemTypeIOS          OperatingSystemType = 2
	OperatingSystemTypeLinux        OperatingSystemType = 3
	OperatingSystemTypeAndroid      OperatingSystemType = 4
	OperatingSystemTypeWindowsPhone OperatingSystemType = 5
)

func (osType OperatingSystemType) String() string {
	switch osType {
	case OperatingSystemTypeWindows:
		return "WINDOWS"
	case OperatingSystemTypeMac:
		return "MAC"
	case OperatingSystemTypeIOS:
		return "IOS"
	case OperatingSystemTypeLinux:
		return "LINUX"
	case OperatingSystemTypeAndroid:
		return "ANDROID"
	case OperatingSystemTypeWindowsPhone:
		return "WINDOWS_PHONE"
	}
	return ""
}

func ParseOperatingSystemType(s string) (OperatingSystemType, bool) {
	switch s {
	case "WINDOWS":
		return OperatingSystemTypeWindows, true
	case "MAC":
		return OperatingSystemTypeMac, true
	case "IOS":
		return OperatingSystemTypeIOS, true
	case "LINUX":
		return OperatingSystemTypeLinux, true
	case "ANDROID":
		return OperatingSystemTypeAndroid, true
	case "WINDOWS_PHONE":
		return OperatingSystemTypeWindowsPhone, true
	}
	return 255, false
}

type OperatingSystem struct {
	duplicationUnsafeSendableBase
	osType OperatingSystemType
}

func NewOperatingSystem(osType OperatingSystemType) *OperatingSystem {
	return &OperatingSystem{osType: osType}
}

func (os OperatingSystem) String() string {
	return fmt.Sprintf("OperatingSystem{osType:%v}", os.osType)
}

func (os *OperatingSystem) dataRestriction() {
	// This method is required to separate external type `Data` from `BaseData` types
}

func (os *OperatingSystem) Type() OperatingSystemType {
	return os.osType
}

func (os *OperatingSystem) QueryEncode() string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPEventType, operatingSystemEventType)
	qb.Append(utils.QPOs, os.osType.String())
	qb.Append(utils.QPOsIndex, fmt.Sprintf("%d", os.osType))
	qb.Append(utils.QPNonce, os.Nonce())
	return qb.String()
}

func (os *OperatingSystem) DataType() DataType {
	return DataTypeOperatingSystem
}
