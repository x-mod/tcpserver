tcpserver
===

## Installation

````
$: go get github.com/x-mod/tcpserver
````

## Quick Start

````
import (
    "net"
    "log"
	"context"
	"github.com/x-mod/tcpserver"
)

func EchoHandler(ctx context.Context, con net.Conn) error {
    //TODO LOGIC
    return nil
}

func main() {
	srv := tcpserver.New(
		tcpserver.Address(":8080"),
		tcpserver.TCPHandler(EchoHandler),
	)
	if err := srv.Serve(context.TODO()); err != nil {
		log.Println("tcpserver failed:", err)
	}
}
````

More Details, Pls check the [example](example/server.go).
