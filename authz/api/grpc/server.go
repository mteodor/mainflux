package grpc

import (
	"context"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/mainflux/mainflux/authz/api"
	pb "github.com/mainflux/mainflux/authz/api/pb"
	"github.com/mainflux/mainflux/errors"
)

type server struct {
	authorize     kitgrpc.Handler
	addThings     kitgrpc.Handler
	addChannels   kitgrpc.Handler
	removeThing   kitgrpc.Handler
	removeChannel kitgrpc.Handler
}

// NewServer returns new AuthZServiceServer instance.
func NewServer(svc api.Service) pb.AuthZServiceServer {
	return &server{
		authorize: kitgrpc.NewServer(
			svc.AuthorizeEndpoint,
			decodeAuthorizeRequest,
			encodeErrorResponse,
		),
		addThings: kitgrpc.NewServer(
			svc.AddThingsEndpoint,
			decodeAddThingsRequest,
			encodeErrorResponse,
		),
		addChannels: kitgrpc.NewServer(
			svc.AddChannelsEndpoint,
			decodeAddChannelsRequest,
			encodeErrorResponse,
		),
		removeChannel: kitgrpc.NewServer(
			svc.RemoveChannelEndpoint,
			decodeRemoveChannelRequest,
			encodeErrorResponse,
		),
		removeThing: kitgrpc.NewServer(
			svc.RemoveThingEndpoint,
			decodeRemoveThingRequest,
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

func (s *server) AddThings(ctx context.Context, req *pb.AddThingsReq) (*pb.ErrorRes, error) {
	_, resp, err := s.addThings.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.ErrorRes), nil
}

func (s *server) AddChannels(ctx context.Context, req *pb.AddChannelsReq) (*pb.ErrorRes, error) {
	_, resp, err := s.addChannels.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.ErrorRes), nil
}

func (s *server) RemoveChannel(ctx context.Context, req *pb.RemoveChannelReq) (*pb.ErrorRes, error) {
	_, resp, err := s.removeChannel.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.ErrorRes), nil
}

func (s *server) RemoveThing(ctx context.Context, req *pb.RemoveThingReq) (*pb.ErrorRes, error) {
	_, resp, err := s.removeThing.ServeGRPC(ctx, req)
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

func decodeAddThingsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.AddThingsReq)
	return api.AddThingsReq{
		Owner: req.GetOwner(),
		IDs:   req.GetIds(),
	}, nil
}

func decodeAddChannelsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.AddChannelsReq)
	return api.AddChannelsReq{
		Owner: req.GetOwner(),
		IDs:   req.GetIds(),
	}, nil
}

func decodeRemoveChannelRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.RemoveChannelReq)
	return api.RemoveChannelReq{
		Owner: req.GetOwner(),
		ID:    req.GetId(),
	}, nil
}

func decodeRemoveThingRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.RemoveThingReq)
	return api.RemoveThingReq{
		Owner: req.GetOwner(),
		ID:    req.GetId(),
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
