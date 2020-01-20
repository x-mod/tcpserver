package tcpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"runtime"
	"sync"

	"golang.org/x/net/trace"
)

//Handler connection handler definition
type Handler func(ctx context.Context, conn net.Conn) error

//Server represent tcpserver
type Server struct {
	name     string
	network  string
	address  string
	handler  Handler
	traced   bool
	mu       sync.Mutex
	events   trace.EventLog
	listener net.Listener
	tlsc     *tls.Config
	wgroup   sync.WaitGroup
}

//Name option for tcpserver
func Name(name string) ServerOpt {
	return func(srv *Server) {
		srv.name = name
	}
}

//Network option for listener
func Network(inet string) ServerOpt {
	return func(srv *Server) {
		if len(inet) != 0 {
			srv.network = inet
		}
	}
}

//Address option for listener
func Address(addr string) ServerOpt {
	return func(srv *Server) {
		if len(addr) != 0 {
			srv.address = addr
		}
	}
}

//TLSConfig option
func TLSConfig(sc *tls.Config) ServerOpt {
	return func(srv *Server) {
		srv.tlsc = sc
	}
}

//Listener option for listener
func Listener(ln net.Listener) ServerOpt {
	return func(srv *Server) {
		if ln != nil {
			srv.listener = ln
		}
	}
}

//TCPHandler option for Connection Handler
func TCPHandler(h Handler) ServerOpt {
	return func(srv *Server) {
		if h != nil {
			srv.handler = h
		}
	}
}

func NetTrace(flag bool) ServerOpt {
	return func(srv *Server) {
		srv.traced = flag
	}
}

//ServerOpt typedef
type ServerOpt func(*Server)

//NewServer create a new tcpserver
func New(opts ...ServerOpt) *Server {
	serv := &Server{
		name:    "tcpserver",
		network: "tcp",
	}
	for _, opt := range opts {
		opt(serv)
	}
	if serv.traced {
		_, file, line, _ := runtime.Caller(1)
		serv.events = trace.NewEventLog(serv.name, fmt.Sprintf("%s:%d", file, line))
	}
	return serv
}

func (srv *Server) printf(format string, a ...interface{}) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.events != nil {
		srv.events.Printf(format, a...)
	}
}

func (srv *Server) errorf(format string, a ...interface{}) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.events != nil {
		srv.events.Errorf(format, a...)
	}
}

//Serve tcpserver serving
func (srv *Server) Serve(ctx context.Context) error {
	if srv.handler == nil {
		return fmt.Errorf("tcpserver.Handler required")
	}
	if srv.listener == nil {
		ln, err := net.Listen(srv.network, srv.address)
		if err != nil {
			return err
		}
		srv.printf("%s serving at %s:%s", srv.name, srv.network, srv.address)
		srv.listener = ln
	}
	if srv.tlsc != nil {
		srv.listener = tls.NewListener(srv.listener, srv.tlsc)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			con, err := srv.listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					srv.errorf("accept temp err: %v", ne)
					continue
				}
				srv.errorf("accept failed: %v", err)
				return err
			}

			srv.wgroup.Add(1)
			go func() {
				defer srv.wgroup.Done()
				if srv.traced {
					tr := trace.New("client", con.RemoteAddr().String())
					ctx = trace.NewContext(ctx, tr)
				}
				if err := srv.handler(ctx, con); err != nil {
					srv.errorf("client (%s) failed: %v", con.RemoteAddr().String(), err)
				}
				if tr, ok := trace.FromContext(ctx); ok {
					tr.Finish()
				}
			}()
		}
	}
}

//Close tcpserver waiting all connections finished
func (srv *Server) Close() {
	_ = srv.listener.Close()
	srv.wgroup.Wait()
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.events != nil {
		srv.events.Finish()
		srv.events = nil
	}
}
