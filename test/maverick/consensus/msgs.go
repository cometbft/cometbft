package consensus

import (
	"errors"
	"fmt"

	cmtcon "github.com/tendermint/tendermint/consensus"
	cstypes "github.com/tendermint/tendermint/consensus/types"
	cmtmath "github.com/tendermint/tendermint/libs/math"
	"github.com/tendermint/tendermint/p2p"
	cmtcons "github.com/tendermint/tendermint/proto/tendermint/consensus"
	cmtproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
)

func WALToProto(msg cmtcon.WALMessage) (*cmtcons.WALMessage, error) {
	var pb cmtcons.WALMessage

	switch msg := msg.(type) {
	case types.EventDataRoundState:
		pb = cmtcons.WALMessage{
			Sum: &cmtcons.WALMessage_EventDataRoundState{
				EventDataRoundState: &cmtproto.EventDataRoundState{
					Height: msg.Height,
					Round:  msg.Round,
					Step:   msg.Step,
				},
			},
		}
	case msgInfo:
		consMsg, err := cmtcon.MsgToProto(msg.Msg)
		if err != nil {
			return nil, err
		}
		pb = cmtcons.WALMessage{
			Sum: &cmtcons.WALMessage_MsgInfo{
				MsgInfo: &cmtcons.MsgInfo{
					Msg:    *consMsg,
					PeerID: string(msg.PeerID),
				},
			},
		}
	case timeoutInfo:
		pb = cmtcons.WALMessage{
			Sum: &cmtcons.WALMessage_TimeoutInfo{
				TimeoutInfo: &cmtcons.TimeoutInfo{
					Duration: msg.Duration,
					Height:   msg.Height,
					Round:    msg.Round,
					Step:     uint32(msg.Step),
				},
			},
		}
	case cmtcon.EndHeightMessage:
		pb = cmtcons.WALMessage{
			Sum: &cmtcons.WALMessage_EndHeight{
				EndHeight: &cmtcons.EndHeight{
					Height: msg.Height,
				},
			},
		}
	default:
		return nil, fmt.Errorf("to proto: wal message not recognized: %T", msg)
	}

	return &pb, nil
}

// WALFromProto takes a proto wal message and return a consensus walMessage and error
func WALFromProto(msg *cmtcons.WALMessage) (cmtcon.WALMessage, error) {
	if msg == nil {
		return nil, errors.New("nil WAL message")
	}
	var pb cmtcon.WALMessage

	switch msg := msg.Sum.(type) {
	case *cmtcons.WALMessage_EventDataRoundState:
		pb = types.EventDataRoundState{
			Height: msg.EventDataRoundState.Height,
			Round:  msg.EventDataRoundState.Round,
			Step:   msg.EventDataRoundState.Step,
		}
	case *cmtcons.WALMessage_MsgInfo:
		walMsg, err := cmtcon.MsgFromProto(&msg.MsgInfo.Msg)
		if err != nil {
			return nil, fmt.Errorf("msgInfo from proto error: %w", err)
		}
		pb = msgInfo{
			Msg:    walMsg,
			PeerID: p2p.ID(msg.MsgInfo.PeerID),
		}

	case *cmtcons.WALMessage_TimeoutInfo:
		tis, err := cmtmath.SafeConvertUint8(int64(msg.TimeoutInfo.Step))
		// deny message based on possible overflow
		if err != nil {
			return nil, fmt.Errorf("denying message due to possible overflow: %w", err)
		}
		pb = timeoutInfo{
			Duration: msg.TimeoutInfo.Duration,
			Height:   msg.TimeoutInfo.Height,
			Round:    msg.TimeoutInfo.Round,
			Step:     cstypes.RoundStepType(tis),
		}
		return pb, nil
	case *cmtcons.WALMessage_EndHeight:
		pb := cmtcon.EndHeightMessage{
			Height: msg.EndHeight.Height,
		}
		return pb, nil
	default:
		return nil, fmt.Errorf("from proto: wal message not recognized: %T", msg)
	}
	return pb, nil
}
