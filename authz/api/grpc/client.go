package grpc

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/mainflux/mainflux"
	pb "github.com/mainflux/mainflux/authz/api/pb"
	"github.com/mainflux/mainflux/pkg/errors"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

const svcName = "authz.AuthZService"

var _ pb.AuthZServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	authorize endpoint.Endpoint
	timeout   time.Duration
}

// NewClient returns new AuthZServiceClient instance.
func NewClient(conn *grpc.ClientConn, tracer opentracing.Tracer) pb.AuthZServiceClient {
	return &grpcClient{
		authorize: kitgrpc.NewClient(
			conn,
			svcName,
			encodeAuthorizeRequest,
			decodeErrorResponse,
			pb.AuthorizeRes{},
		).Endpoint(),
		timeout: timeout,
	}

}

func (client grpcClient) Authorize(ctx context.Context, req *pb.AuthorizeReq, _ ...grpc.CallOption) (bool, error) {
	ctx, close := context.WithTimeout(ctx, client.timeout)
	defer close()

	res, err := client.authorize(ctx, AuthZReq{Act: req.Act, Obj: r })
	if err != nil {
		return nil, err
	}

	ir := res.(identityRes)
	return &mainflux.Token{Value: ir.id}, ir.err
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
