package client

import (
	"context"
	"time"

	"github.com/karlsen-network/karlsend/cmd/karlsenwallet/daemon/server"

	"github.com/pkg/errors"

	"github.com/karlsen-network/karlsend/cmd/karlsenwallet/daemon/pb"
	"google.golang.org/grpc"
)

// Connect connects to the karlsenwalletd server, and returns the client instance
func Connect(address string) (pb.KarlsenwalletdClient, func(), error) {
	// Connection is local, so 1 second timeout is sufficient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(server.MaxDaemonSendMsgSize)))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, errors.New("karlsenwallet daemon is not running, start it with `karlsenwallet start-daemon`")
		}
		return nil, nil, err
	}

	return pb.NewKarlsenwalletdClient(conn), func() {
		conn.Close()
	}, nil
}
