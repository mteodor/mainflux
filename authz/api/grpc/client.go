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

}

func encodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(api.AuthZReq)
	return &pb.AuthorizeReq{
		Sub: req.Sub,
		Obj: req.Obj,
		Act: req.Act,
	}, nil
}

func decodeErrorResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*pb.ErrorRes)
	return api.ErrorRes{
		Err: errors.New(res.Err),
	}, nil
}
