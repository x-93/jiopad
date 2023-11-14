package rpchandlers

import (
	"github.com/karlsen-network/karlsend/app/appmessage"
	"github.com/karlsen-network/karlsend/app/rpc/rpccontext"
	"github.com/karlsen-network/karlsend/infrastructure/network/netadapter/router"
)

// HandleGetHeaders handles the respectively named RPC command
func HandleGetHeaders(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	response := &appmessage.GetHeadersResponseMessage{}
	response.Error = appmessage.RPCErrorf("not implemented")
	return response, nil
}
