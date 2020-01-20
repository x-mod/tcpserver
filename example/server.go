package main

import (
	"context"
	"log"
	"net"
	"net/textproto"
	"strings"

	"golang.org/x/net/trace"

	"github.com/x-mod/routine"
	"github.com/x-mod/tcpserver"
)

func main() {
	srv := tcpserver.New(
		tcpserver.Network("tcp"),
		tcpserver.Address("127.0.0.1:8080"),
		tcpserver.TCPHandler(EchoHandler),
		tcpserver.NetTrace(true),
	)
	log.Println("tcpserver serving ...")
	if err := routine.Main(
		context.TODO(),
		routine.ExecutorFunc(srv.Serve),
		routine.Go(routine.Profiling("127.0.0.1:6060")),
	); err != nil {
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
			if tr, ok := trace.FromContext(ctx); ok {
				tr.LazyPrintf("echo request: %s", line)
			}
		}
	}

}
