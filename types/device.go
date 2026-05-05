package types

import (
	"fmt"

	"github.com/Kameleoon/client-go/v3/utils"
)

type DeviceType string

const (
	DeviceTypeDesktop DeviceType = "DESKTOP"
	DeviceTypePhone   DeviceType = "PHONE"
	DeviceTypeTablet  DeviceType = "TABLET"
)

const deviceEventType = "staticData"

type Device struct {
	duplicationUnsafeSendableBase
	deviceType DeviceType
}

func NewDevice(deviceType DeviceType) *Device {
	return &Device{deviceType: deviceType}
}

func (d Device) String() string {
	return fmt.Sprintf("Device{deviceType:%v}", d.deviceType)
}

func (d *Device) dataRestriction() {}

func (d *Device) Type() DeviceType {
	return d.deviceType
}

func (d *Device) QueryEncode() string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPEventType, deviceEventType)
	qb.Append(utils.QPDeviceType, string(d.deviceType))
	qb.Append(utils.QPNonce, d.Nonce())
	return qb.String()
}

func (d *Device) DataType() DataType {
	return DataTypeDevice
}
