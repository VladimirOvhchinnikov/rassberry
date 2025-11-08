package grpc

import (
	"context"
	"net"
	"sync"
)

type ServerOption func(*serverOptions)

type serverOptions struct {
	streamInterceptor StreamServerInterceptor
}

type Server struct {
	opts     serverOptions
	stopCh   chan struct{}
	stopOnce sync.Once
}

const SupportPackageIsVersion7 = true

func NewServer(opts ...ServerOption) *Server {
	s := &Server{stopCh: make(chan struct{})}
	for _, opt := range opts {
		if opt != nil {
			opt(&s.opts)
		}
	}
	return s
}

func (s *Server) RegisterService(_ *ServiceDesc, _ interface{}) {}

func (s *Server) Serve(lis net.Listener) error {
	if lis != nil {
		go func() {
			<-s.stopCh
			_ = lis.Close()
		}()
	}
	<-s.stopCh
	return nil
}

func (s *Server) GracefulStop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
}

type ServiceRegistrar interface {
	RegisterService(*ServiceDesc, interface{})
}

type StreamHandler func(srv interface{}, stream ServerStream) error

type StreamServerInfo struct {
	FullMethod     string
	IsClientStream bool
	IsServerStream bool
}

type StreamServerInterceptor func(srv interface{}, ss ServerStream, info *StreamServerInfo, handler StreamHandler) error

func StreamInterceptor(i StreamServerInterceptor) ServerOption {
	return func(o *serverOptions) {
		o.streamInterceptor = i
	}
}

type ServerStream interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
}

type StreamDesc struct {
	StreamName    string
	Handler       StreamHandler
	ServerStreams bool
	ClientStreams bool
}

type ServiceDesc struct {
	ServiceName string
	HandlerType interface{}
	Streams     []StreamDesc
	Metadata    interface{}
}

var _ ServiceRegistrar = (*Server)(nil)
