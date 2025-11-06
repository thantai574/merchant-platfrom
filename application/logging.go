package application

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func (us *OrderApplication) WrapperLoggingGRPC() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, _ := metadata.FromIncomingContext(ctx)

		us.Logger.With(
			zap.Field{
				Key:       "request",
				Interface: req,
				Type:      zapcore.ReflectType,
			},

			zap.Field{
				Key:       "info-server",
				Interface: info.FullMethod,
				Type:      zapcore.ReflectType,
			},

			zap.Field{
				Key:       "md-value",
				Interface: md,
				Type:      zapcore.ReflectType,
			},

			zap.Field{
				Key:       "ctx",
				Interface: ctx,
				Type:      zapcore.ReflectType,
			},
		).Info("info")
		return handler(ctx, req)
	}
}
