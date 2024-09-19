package server

import (
	"context"

	"github.com/karlsen-network/karlsend/v2/cmd/karlsenwallet/daemon/pb"
	"github.com/karlsen-network/karlsend/v2/version"
)

func (s *server) GetVersion(_ context.Context, _ *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return &pb.GetVersionResponse{
		Version: version.Version(),
	}, nil
}
