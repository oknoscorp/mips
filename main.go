package main

import (
	"context"
	"os"

	"github.com/oknoscorp/mips/pkg/polling"
	"github.com/oknoscorp/mips/pkg/server"
)

func main() {

	// Initialize TCP server.
	serv := server.New()

	// Initialize Worker related polling component.
	polling.New(serv)

	ctx := context.Background()
	select {
	case <-ctx.Done():
		os.Exit(1)
	}

}
