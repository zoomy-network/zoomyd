package blockrelay

import (
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
	peerpkg "github.com/zoomy-network/zoomyd/app/protocol/peer"
	"github.com/zoomy-network/zoomyd/app/protocol/protocolerrors"
	"github.com/zoomy-network/zoomyd/domain"
	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter/router"
)

// RelayBlockRequestsContext is the interface for the context needed for the HandleRelayBlockRequests flow.
type RelayBlockRequestsContext interface {
	Domain() domain.Domain
}

// HandleRelayBlockRequests listens to appmessage.MsgRequestRelayBlocks messages and sends
// their corresponding blocks to the requesting peer.
func HandleRelayBlockRequests(context RelayBlockRequestsContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peerpkg.Peer) error {

	for {
		message, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		getRelayBlocksMessage := message.(*appmessage.MsgRequestRelayBlocks)
		log.Debugf("Got request for relay blocks with hashes %s", getRelayBlocksMessage.Hashes)
		for _, hash := range getRelayBlocksMessage.Hashes {
			// Fetch the block from the database.
			block, found, err := context.Domain().Consensus().GetBlock(hash)
			if err != nil {
				return errors.Wrapf(err, "unable to fetch requested block hash %s", hash)
			}

			if !found {
				return protocolerrors.Errorf(false, "Relay block %s not found", hash)
			}

			// TODO (Partial nodes): Convert block to partial block if needed

			err = outgoingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(block))
			if err != nil {
				return err
			}
			log.Debugf("Relayed block with hash %s", hash)
		}
	}
}
