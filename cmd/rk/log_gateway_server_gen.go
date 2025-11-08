package main

import (
	"context"
	"io"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"example.com/ffp/platform/telemetry"
	telemetrypb "example.com/ffp/platform/telemetry/proto"
)

type logGatewayServer struct {
	telemetrypb.UnimplementedLogGatewayServer
	hub         *LogHub
	ackInterval time.Duration
}

func (s *logGatewayServer) PushLogs(stream telemetrypb.LogGateway_PushLogsServer) error {
	var count uint64
	ctx := stream.Context()
	last := time.Now()

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			_ = stream.Send(&telemetrypb.PushResponse{Status: "ok", Received: count})
			return nil
		}
		if err != nil {
			return err
		}
		rec := telemetry.FromProto(msg)
		s.hub.Publish(ctx, rec)
		count++

		if time.Since(last) >= s.ackInterval {
			_ = stream.Send(&telemetrypb.PushResponse{Status: "ok", Received: count})
			last = time.Now()
		}
	}
}

func StartLogGatewayServer(ctx context.Context, addr string, hub *LogHub) error {
	if addr == "" {
		addr = ":8079"
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer(
		grpc.StreamInterceptor(func(
			srv any,
			ss grpc.ServerStream,
			info *grpc.StreamServerInfo,
			handler grpc.StreamHandler,
		) error {
			// легкий лог: кто подключился
			if p, ok := peer.FromContext(ss.Context()); ok && p.Addr != nil {
				_ = p.Addr.String()
			}
			return handler(srv, ss)
		}),
	)
	telemetrypb.RegisterLogGatewayServer(srv, &logGatewayServer{hub: hub, ackInterval: time.Second})

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	return srv.Serve(lis)
}
