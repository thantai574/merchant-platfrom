package main

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/micro/go-micro/v2/metadata"
	_ "github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"net"
	"orders-system/application"
	"orders-system/presenters"
	"orders-system/proto/order_system"
	"orders-system/utils/configs"
	"orders-system/utils/context_grpc"
	"orders-system/utils/errors"
	"orders-system/utils/gen_ids"
	_ "orders-system/utils/gen_ids"
	"orders-system/utils/gpooling"
	logger2 "orders-system/utils/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Create service
	// Register handler
	config, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}
	lg, _ := logger2.NewLogger("production")

	pool_go_routine, _ := gpooling.NewPooling(config.MaxPoolSize)

	app := application.NewOrderApplication(config, lg, pool_go_routine)

	gen_ids.InitGenIDservice()

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_opentracing.UnaryServerInterceptor(),
				grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
					lg.With(zap.Field{
						Key:       "context",
						Interface: p,
						Type:      zapcore.ReflectType,
					}).Error("error")

					return errors.RecoveryError(p)
				})),
				func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
					new_ctx := context_grpc.NewOrderSystemContextGRPC(ctx)
					new_ctx.StartSpan(ctx, info.FullMethod)
					ctx = new_ctx

					logs := lg.With(
						zap.Field{
							Key:       "request",
							Interface: req,
							Type:      zapcore.ReflectType,
						},
					)

					resp, err = handler(ctx, req)
					new_ctx.Duration()
					meta_data_from_context, _ := metadata.FromContext(ctx)

					if err != nil {
						err = errors.RecoveryError(err)
						logs.With(zap.Field{
							Key:       "metadata",
							Interface: meta_data_from_context,
							Type:      zapcore.ReflectType,
						}, zap.Field{
							Key:       "error-context",
							Interface: err,
							Type:      zapcore.ReflectType,
						}, zap.Field{
							Key:       "trace-id",
							Interface: new_ctx.TraceId,
							Type:      zapcore.ReflectType,
						}, zap.Field{
							Key:       "method",
							Interface: info.FullMethod,
							Type:      zapcore.ReflectType,
						}).Error("response-error")
					} else {
						logs.With(zap.Field{
							Key:       "metadata",
							Interface: meta_data_from_context,
							Type:      zapcore.ReflectType,
						}, zap.Field{
							Key:       "response-context",
							Interface: resp,
							Type:      zapcore.ReflectType,
						}, zap.Field{
							Key:       "trace-id",
							Interface: new_ctx.TraceId,
							Type:      zapcore.ReflectType,
						}, zap.Field{
							Key:       "method",
							Interface: info.FullMethod,
							Type:      zapcore.ReflectType,
						}).Info("response-success")
					}

					return
				},
			)),
	)

	order_system.RegisterOrdersSystemServer(srv, presenters.NewOrderSystemGRPC(app))

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)

	pool_go_routine.Submit(func() {
		func() {
			select {
			case <-sig:
				lg.Warn("shutting down gRPC server...")
				srv.GracefulStop()
				pool_go_routine.Release()
			}
		}()
	})

	if config.Job {
		pool_go_routine.Submit(func() {
			for {
				select {
				case <-time.Tick(time.Second * 3):
					app.JobLixi()
				}
			}
		})
	}

	go func() {
		pool_go_routine.Submit(func() {
			for {
				select {
				case <-time.Tick(time.Second * 5):
					app.JobCancelExpiredOrder()
				}
			}
		})
	}()

	lis, err := net.Listen("tcp", ":"+config.Port)
	// Run service
	lg.With(zap.Field{
		Key:    "port",
		Type:   zapcore.StringType,
		String: config.Port,
	}).Info("starting gRPC serverv1...")

	if err != nil {
		panic(err)
	}
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
