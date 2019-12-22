package tcpserver

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"sync"
)

//Handler connection handler definition
type Handler func(ctx context.Context, conn net.Conn) error

//Server represent tcpserver
type Server struct {
	name     string
	network  string
	address  string
	handler  Handler
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
	return serv
}

//Serve tcpserver serving
func (srv *Server) Serve(ctx context.Context) error {
	if srv.listener == nil {
		ln, err := net.Listen(srv.network, srv.address)
		if err != nil {
			return err
		}
		log.Println(srv.name, " serving at ", srv.network, srv.address)
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
					log.Printf("warning: accept temp err: %v", ne)
					continue
				}
				log.Println("failed: ", err)
				return err
			}

			srv.wgroup.Add(1)
			go func() {
				defer srv.wgroup.Done()
				if err := srv.handler(ctx, con); err != nil {
					log.Printf("connection %s handle failed: %s\n", con.RemoteAddr().String(), err)
				}
			}()
		}
	}
}

//Close tcpserver waiting all connections finished
func (srv *Server) Close() {
	_ = srv.listener.Close()
	srv.wgroup.Wait()
}
