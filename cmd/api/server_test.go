package main_test

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	"orders-system/application/test"
	"orders-system/presenters"
	"orders-system/proto/order_system"
	"orders-system/utils/errors"
	logger2 "orders-system/utils/logger"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func dialer() func(context.Context, string) (net.Conn, error) {
	lg, _ := logger2.NewLogger("production")
	app := test.NewTestOrderApplication()

	listener := bufconn.Listen(1024 * 1024)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			app.OrderApplication.WrapperLoggingGRPC(),
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {

				lg.With(zap.Field{
					Key:       "context",
					Interface: p,
					Type:      zapcore.ReflectType,
				}).Error("error")

				return errors.RecoveryError(p)
			})),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				resp, err = handler(ctx, req)

				if err != nil {
					err = errors.RecoveryError(err)
				}

				return
			},
		)),
	)

	order_system.RegisterOrdersSystemServer(srv, presenters.NewOrderSystemGRPC(app.OrderApplication))

	go func() {
		if err := srv.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestDepositServer_Call(t *testing.T) {

	tests := []struct {
		name    string
		amount  float32
		errCode codes.Code
		errMsg  string
	}{
		{
			"invalid request with negative amount",
			-1.11,
			codes.InvalidArgument,
			fmt.Sprintf("cannot deposit %v", -1.11),
		},
	}

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithAuthority("dialer()"), grpc.WithContextDialer(dialer()))
	//conn, err := grpc.Dial("localhost:10001", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := order_system.NewOrdersSystemClient(conn)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := metadata.Pairs("s", "v1", "k2", "v2", "k3", "v3")
			_, err = client.BuyCard(metadata.NewOutgoingContext(ctx, md), &order_system.OrderBuyCardRequest{
				Telco: "VTM",
				OrderRequest: &order_system.OrderRequest{
					Amount:               10000,
					Quantity:             1,
					VoucherCode:          "",
					UserID:               "US547856US5F7166F16E536949308",
					MerchantID:           "",
					ServiceID:            "BUYCARD",
					SubTransType:         "BUYCARD",
					TransType:            "BALANCE_WAL",
					DeviceID:             "",
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			})

			if err != nil {
				if er, ok := status.FromError(err); ok {
					t.Log(er)
					//if er.Code() != tt.errCode {
					//	t.Error("error code: expected", codes.InvalidArgument, "received", er.Code())
					//}
					//if er.Message() != tt.errMsg {
					//	t.Error("error message: expected", tt.errMsg, "received", er.Message())
					//}
				}
			}
		})
	}
}
