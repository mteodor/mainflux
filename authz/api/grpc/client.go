package grpc

import (
	"context"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/mainflux/mainflux/authz"
	"github.com/mainflux/mainflux/authz/api"
	pb "github.com/mainflux/mainflux/authz/api/pb"
	"github.com/mainflux/mainflux/errors"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

const svcName = "authz.AuthZService"

// NewClient returns new AuthZServiceClient instance.
func NewClient(conn *grpc.ClientConn, tracer opentracing.Tracer) authz.Service {
	authorize := kitgrpc.NewClient(
		conn,
		svcName,
		"Authorize",
		encodeAuthorizeRequest,
		decodeErrorResponse,
		pb.ErrorRes{},
	).Endpoint()
	authorize = kitot.TraceClient(tracer, "authorize")(authorize)

	addThings := kitgrpc.NewClient(
		conn,
		svcName,
		"AddThings",
		encodeAddThingsRequest,
		decodeErrorResponse,
		pb.ErrorRes{},
	).Endpoint()
	addThings = kitot.TraceClient(tracer, "add_things")(addThings)

	addChannels := kitgrpc.NewClient(
		conn,
		svcName,
		"AddChannels",
		encodeAddChannelsRequest,
		decodeErrorResponse,
		pb.ErrorRes{},
	).Endpoint()
	addChannels = kitot.TraceClient(tracer, "add_channels")(addChannels)

	removeChannel := kitgrpc.NewClient(
		conn,
		svcName,
		"RemoveChannel",
		encodeRemoveChannelRequest,
		decodeErrorResponse,
		pb.ErrorRes{},
	).Endpoint()
	removeChannel = kitot.TraceClient(tracer, "remove_channel")(removeChannel)

	removeThing := kitgrpc.NewClient(
		conn,
		svcName,
		"RemoveThing",
		encodeRemoveThingRequest,
		decodeErrorResponse,
		pb.ErrorRes{},
	).Endpoint()

	return api.Service{
		AuthorizeEndpoint:     authorize,
		AddThingsEndpoint:     addThings,
		AddChannelsEndpoint:   addChannels,
		RemoveChannelEndpoint: removeChannel,
		RemoveThingEndpoint:   removeThing,
	}
}

func encodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(api.AuthZReq)
	return &pb.AuthorizeReq{
		Sub: req.Sub,
		Obj: req.Obj,
		Act: req.Act,
	}, nil
}

func encodeAddThingsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(api.AddThingsReq)
	return &pb.AddThingsReq{
		Owner: req.Owner,
		Ids:   req.IDs,
	}, nil
}

func encodeAddChannelsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(api.AddChannelsReq)
	return &pb.AddChannelsReq{
		Owner: req.Owner,
		Ids:   req.IDs,
	}, nil
}

func encodeRemoveChannelRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(api.RemoveChannelReq)
	return &pb.RemoveChannelReq{
		Owner: req.Owner,
		Id:    req.ID,
	}, nil
}

func encodeRemoveThingRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(api.RemoveThingReq)
	return &pb.RemoveThingReq{
		Owner: req.Owner,
		Id:    req.ID,
	}, nil
}

func decodeErrorResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*pb.ErrorRes)
	return api.ErrorRes{
		Err: errors.New(res.Err),
	}, nil
}
