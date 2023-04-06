package consensus

import (
	"github.com/cometbft/cometbft/api/cometbft/consensus/v1"
	"github.com/cometbft/cometbft/api/cometbft/consensus/v2"
)

type WALMessage = v2.WALMessage
type WALMessage_EndHeight = v2.WALMessage_EndHeight
type WALMessage_EventDataRoundState = v2.WALMessage_EventDataRoundState
type WALMessage_MsgInfo = v2.WALMessage_MsgInfo
type WALMessage_TimeoutInfo = v2.WALMessage_TimeoutInfo

type EndHeight = v1.EndHeight
type MsgInfo = v2.MsgInfo
type TimeoutInfo = v1.TimeoutInfo

type TimedWALMessage = v2.TimedWALMessage
