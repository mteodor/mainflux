package grpc

import (
	"context"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/mainflux/mainflux/authz/api"
	pb "github.com/mainflux/mainflux/authz/api/pb"
	"github.com/mainflux/mainflux/errors"
)

type server struct {
	authorize kitgrpc.Handler
}

// NewServer returns new AuthZServiceServer instance.
func NewServer(svc api.Service) pb.AuthZServiceServer {
	return &server{
		authorize: kitgrpc.NewServer(
			svc.AuthorizeEndpoint,
			decodeAuthorizeRequest,
			encodeErrorResponse,
		),
	}
}

func (s *server) Authorize(ctx context.Context, req *pb.AuthorizeReq) (*pb.ErrorRes, error) {
	_, resp, err := s.authorize.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.ErrorRes), nil
}

func decodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.AuthorizeReq)
	return api.AuthZReq{
		Sub: req.GetSub(),
		Obj: req.GetObj(),
		Act: req.GetAct(),
	}, nil
}

func encodeErrorResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(api.ErrorRes)
	return &pb.ErrorRes{
		Err: err2str(res.Err),
	}, nil
}

func err2str(err error) string {
	if err != nil {
		e, ok := err.(errors.Error)
		if !ok {
			return err.Error()
		}

		return e.Msg()
	}

	return ""
}
