package main

import (
	"context"
	"log"
	"net"
	"net/textproto"
	"strings"
	"syscall"

	"github.com/x-mod/glog"
	"github.com/x-mod/routine"
	"github.com/x-mod/tcpserver"
	"golang.org/x/net/trace"
)

func main() {
	glog.Open(
		glog.LogToStderr(true),
		glog.Verbosity(5),
	)
	defer glog.Close()
	srv := tcpserver.New(
		tcpserver.Network("tcp"),
		tcpserver.Address("127.0.0.1:8080"),
		tcpserver.TCPHandler(EchoHandler),
		tcpserver.NetTrace(true),
		// tcpserver.TLSConfig(tlsconfig.New(
		// 	tlsconfig.CertKeyPair("out/server.crt", "out/server.key"),
		// )),
	)
	log.Println("tcpserver serving ...")
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	if err := routine.Main(
		ctx,
		routine.ExecutorFunc(func(ctx context.Context) error {
			<-srv.Serving()
			log.Println("serving ... now")
			<-ctx.Done()
			return nil
		}),
		routine.Go(routine.ExecutorFunc(srv.Serve)),
		routine.Go(routine.Profiling("127.0.0.1:6060")),
		routine.Signal(syscall.SIGINT, routine.SigHandler(func() {
			cancel()
			<-srv.Close()
		})),
	); err != nil {
		log.Println("tcpserver failed:", err)
	}
}

func EchoHandler(ctx context.Context, con net.Conn) error {
	defer con.Close()
	c := textproto.NewConn(con)

	c.PrintfLine("HELLO From Server")
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
