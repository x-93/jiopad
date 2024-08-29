package protowire

import (
	"github.com/karlsen-network/karlsend/v2/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KarlsendMessage_Ready) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KarlsendMessage_Ready is nil")
	}
	return &appmessage.MsgReady{}, nil
}

func (x *KarlsendMessage_Ready) fromAppMessage(_ *appmessage.MsgReady) error {
	return nil
}
