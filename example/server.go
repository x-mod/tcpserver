package main

import (
	"context"
	"log"
	"net"
	"net/textproto"
	"strings"

	"github.com/x-mod/routine"
	"github.com/x-mod/tcpserver"
)

func main() {
	srv := tcpserver.NewServer(
		tcpserver.Network("tcp"),
		tcpserver.Address("127.0.0.1:8080"),
		tcpserver.Handler(EchoHandler),
	)
	if err := routine.Main(routine.ExecutorFunc(srv.Serve)); err != nil {
		log.Println("tcpserver failed:", err)
	}
}

func EchoHandler(ctx context.Context, con net.Conn) error {
	defer con.Close()

	c := textproto.NewConn(con)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line, err := c.ReadLine()
			if err != nil {
				return err
			}
			if strings.HasPrefix(line, "quit") {
				return nil
			}
			if err := c.PrintfLine(line); err != nil {
				return err
			}
		}
	}
}
