package types

import "github.com/Kameleoon/client-go/v3/utils"

type Sendable interface {
	QueryEncodable

	Nonce() string

	Unsent() bool
	MarkAsUnsent()

	Transmitting() bool
	MarkAsTransmitting()

	Sent() bool
	MarkAsSent()
}

type SendableState byte

const (
	SendableStateUnsent       SendableState = 0
	SendableStateTransmitting SendableState = 1
	SendableStateSent         SendableState = 2
)

type sendableBase struct {
	nonce string
	state SendableState
}

func (sb *sendableBase) Unsent() bool {
	return sb.state == SendableStateUnsent
}
func (sb *sendableBase) MarkAsUnsent() {
	if sb.Transmitting() {
		sb.state = SendableStateUnsent
	}
}

func (sb *sendableBase) Transmitting() bool {
	return sb.state == SendableStateTransmitting
}
func (sb *sendableBase) MarkAsTransmitting() {
	if sb.Unsent() {
		sb.state = SendableStateTransmitting
	}
}

func (sb *sendableBase) Sent() bool {
	return sb.state == SendableStateSent
}
func (sb *sendableBase) MarkAsSent() {
	sb.state = SendableStateSent
	sb.nonce = ""
}

type duplicationSafeSendableBase struct {
	sendableBase
}

func (sb *duplicationSafeSendableBase) initSendable() {
	sb.nonce = utils.GetNonce()
}

func (sb *duplicationSafeSendableBase) Nonce() string {
	return sb.nonce
}

type duplicationUnsafeSendableBase struct {
	sendableBase
}

func (sb *duplicationUnsafeSendableBase) Nonce() string {
	if !sb.Sent() && len(sb.nonce) == 0 {
		sb.nonce = utils.GetNonce()
	}
	return sb.nonce
}
