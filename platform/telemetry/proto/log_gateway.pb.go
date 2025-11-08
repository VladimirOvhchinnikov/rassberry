package proto

import (
	"fmt"

	"google.golang.org/grpc"
)

type PushRequest struct {
	Record *LogRecord
}

type PushResponse struct {
	Status   string
	Received uint64
}

type LogRecord struct {
	TimeUnixNano int64
	Level        string
	KernelId     string
	Scope        string
	Component    string
	Trace        string
	Message      string
	Fields       []*LogRecordField
}

type LogRecordField struct {
	Key   string
	Value string
}

type LogGatewayServer interface {
	PushLogs(LogGateway_PushLogsServer) error
}

type UnimplementedLogGatewayServer struct{}

func (UnimplementedLogGatewayServer) PushLogs(LogGateway_PushLogsServer) error {
	return fmt.Errorf("method PushLogs not implemented")
}

func RegisterLogGatewayServer(s grpc.ServiceRegistrar, srv LogGatewayServer) {
	s.RegisterService(&LogGateway_ServiceDesc, srv)
}

type LogGateway_PushLogsServer interface {
	Send(*PushResponse) error
	Recv() (*PushRequest, error)
	grpc.ServerStream
}

func _LogGateway_PushLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LogGatewayServer).PushLogs(&logGatewayPushLogsServer{stream})
}

type logGatewayPushLogsServer struct {
	grpc.ServerStream
}

func (x *logGatewayPushLogsServer) Send(m *PushResponse) error { return x.ServerStream.SendMsg(m) }

func (x *logGatewayPushLogsServer) Recv() (*PushRequest, error) {
	req := new(PushRequest)
	if err := x.ServerStream.RecvMsg(req); err != nil {
		return nil, err
	}
	return req, nil
}

var LogGateway_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "telemetry.LogGateway",
	HandlerType: (*LogGatewayServer)(nil),
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PushLogs",
			Handler:       _LogGateway_PushLogs_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
}
