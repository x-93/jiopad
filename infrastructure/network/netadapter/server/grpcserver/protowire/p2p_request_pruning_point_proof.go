package protowire

import (
	"github.com/karlsen-network/karlsend/v2/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KarlsendMessage_RequestPruningPointProof) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KarlsendMessage_RequestPruningPointProof is nil")
	}
	return &appmessage.MsgRequestPruningPointProof{}, nil
}

func (x *KarlsendMessage_RequestPruningPointProof) fromAppMessage(_ *appmessage.MsgRequestPruningPointProof) error {
	return nil
}
