package blocksync

import "github.com/cometbft/cometbft/p2p"

func blockQueueFull(peer p2p.Peer) bool {
	if peer == nil {
		return false
	}
	for _, channel := range peer.Status().Channels {
		if channel.ID == BlocksyncChannel {
			return channel.SendQueueCapacity > 0 && channel.SendQueueSize >= channel.SendQueueCapacity
		}
	}
	return false
}
