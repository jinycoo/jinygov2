package warden

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"go.jd100.com/medusa/errors"
	"go.jd100.com/medusa/log"
	"go.jd100.com/medusa/net/metadata"
	"go.jd100.com/medusa/stat"
)

var (
	statsClient = stat.RPCClient
	statsServer = stat.RPCServer
)

func logFn(code int, dt time.Duration) func(context.Context, string, ...log.Field) {
	switch {
	case code < 0:
		return log.Errorw
	case dt >= time.Millisecond*500:
		// TODO: slowlog make it configurable.
		return log.Warnw
	case code > 0:
		return log.Warnw
	}
	return log.Infow
}

// clientLogging warden grpc logging
func clientLogging() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()
		var peerInfo peer.Peer
		opts = append(opts, grpc.Peer(&peerInfo))

		// invoker requests
		err := invoker(ctx, method, req, reply, cc, opts...)

		// after request
		code := errors.ECause(err).Code()
		duration := time.Since(startTime)
		// monitor
		statsClient.Timing(method, int64(duration/time.Millisecond))
		statsClient.Incr(method, strconv.Itoa(code))

		var ip string
		if peerInfo.Addr != nil {
			ip = peerInfo.Addr.String()
		}
		logFields := []log.Field{
			log.String("ip", ip),
			log.String("path", method),
			log.Int("ret", code),
			// TODO: it will panic if someone remove String method from protobuf message struct that auto generate from protoc.
			log.String("args", req.(fmt.Stringer).String()),
			log.Float64("ts", duration.Seconds()),
			log.String("source", "grpc-access-log"),
		}
		if err != nil {
			logFields = append(logFields, log.String("error", err.Error()), log.String("stack", fmt.Sprintf("%+v", err)))
		}
		logFn(code, duration)(ctx, "", logFields...)
		return err
	}
}

// serverLogging warden grpc logging
func serverLogging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		caller := metadata.String(ctx, metadata.Caller)
		var remoteIP string
		if peerInfo, ok := peer.FromContext(ctx); ok {
			remoteIP = peerInfo.Addr.String()
		}
		var quota float64
		if deadline, ok := ctx.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		// call server handler
		resp, err := handler(ctx, req)

		// after server response
		code := errors.ECause(err).Code()
		duration := time.Since(startTime)

		// monitor
		statsServer.Timing(caller, int64(duration/time.Millisecond), info.FullMethod)
		statsServer.Incr(caller, info.FullMethod, strconv.Itoa(code))
		logFields := []log.Field{
			log.String("user", caller),
			log.String("ip", remoteIP),
			log.String("path", info.FullMethod),
			log.Int("ret", code),
			// TODO: it will panic if someone remove String method from protobuf message struct that auto generate from protoc.
			log.String("args", req.(fmt.Stringer).String()),
			log.Float64("ts", duration.Seconds()),
			log.Float64("timeout_quota", quota),
			log.String("source", "grpc-access-log"),
		}
		if err != nil {
			logFields = append(logFields, log.String("error", err.Error()), log.String("stack", fmt.Sprintf("%+v", err)))
		}
		logFn(code, duration)(ctx, "", logFields...)
		return resp, err
	}
}
