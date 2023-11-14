package protowire

import (
	"github.com/karlsen-network/karlsend/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KarlsendMessage_Verack) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KarlsendMessage_Verack is nil")
	}
	return &appmessage.MsgVerAck{}, nil
}

func (x *KarlsendMessage_Verack) fromAppMessage(_ *appmessage.MsgVerAck) error {
	return nil
}
