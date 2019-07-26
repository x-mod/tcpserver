package tcpserver

import (
	"context"
	"log"
	"net"
)

//ConnectionHandler connection handler definition
type ConnectionHandler func(ctx context.Context, conn net.Conn) error

//Server represent tcpserver
type Server struct {
	network string
	address string
	handler ConnectionHandler
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

//Handler option for connection
func Handler(h ConnectionHandler) ServerOpt {
	return func(srv *Server) {
		if h != nil {
			srv.handler = h
		}
	}
}

//ServerOpt typedef
type ServerOpt func(*Server)

//NewServer create a new tcpserver
func NewServer(opts ...ServerOpt) *Server {
	serv := &Server{}
	for _, opt := range opts {
		opt(serv)
	}
	return serv
}

//Serve tcpserver serving
func (srv *Server) Serve(ctx context.Context) error {
	ln, err := net.Listen(srv.network, srv.address)
	if err != nil {
		return err
	}
	log.Println("tcpserver serving at ", srv.network, srv.address)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			con, err := ln.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					log.Printf("warning: accept temp err: %v", ne)
					continue
				}
				return err
			}

			go func() {
				if err := srv.handler(ctx, con); err != nil {
					log.Printf("connection %s handle failed: %s\n", con.RemoteAddr().String(), err)
				}
			}()
		}
	}
}
