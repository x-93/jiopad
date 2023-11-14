package protowire

import (
	"github.com/karlsen-network/karlsend/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KarlsendMessage_RequestNextPruningPointUtxoSetChunk) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KarlsendMessage_RequestNextPruningPointUtxoSetChunk is nil")
	}
	return &appmessage.MsgRequestNextPruningPointUTXOSetChunk{}, nil
}

func (x *KarlsendMessage_RequestNextPruningPointUtxoSetChunk) fromAppMessage(_ *appmessage.MsgRequestNextPruningPointUTXOSetChunk) error {
	x.RequestNextPruningPointUtxoSetChunk = &RequestNextPruningPointUtxoSetChunkMessage{}
	return nil
}
