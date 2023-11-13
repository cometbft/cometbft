//nolint:revive,stylecheck
package proto

import (
	"github.com/cometbft/cometbft/api/cometbft/consensus/v1beta1"
	"github.com/cometbft/cometbft/api/cometbft/consensus/v1beta2"
)

type (
	WALMessage                     = v1beta2.WALMessage
	WALMessage_EndHeight           = v1beta2.WALMessage_EndHeight
	WALMessage_EventDataRoundState = v1beta2.WALMessage_EventDataRoundState
	WALMessage_MsgInfo             = v1beta2.WALMessage_MsgInfo
	WALMessage_TimeoutInfo         = v1beta2.WALMessage_TimeoutInfo
)

type (
	EndHeight   = v1beta1.EndHeight
	MsgInfo     = v1beta2.MsgInfo
	TimeoutInfo = v1beta1.TimeoutInfo
)

type TimedWALMessage = v1beta2.TimedWALMessage
