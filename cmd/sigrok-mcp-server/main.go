package main

import (
	"fmt"
	"os"

	"github.com/KenosInc/sigrok-mcp-server/internal/config"
	"github.com/KenosInc/sigrok-mcp-server/internal/sigrok"
	"github.com/KenosInc/sigrok-mcp-server/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	cfg := config.Load()
	executor := sigrok.NewExecutor(cfg.SigrokCLIPath, cfg.Timeout, cfg.WorkingDir)
	handlers := tools.NewHandlers(executor)

	srv := server.NewMCPServer("sigrok-mcp-server", "0.1.0")
	tools.RegisterAll(srv, handlers)

	if err := server.ServeStdio(srv); err != nil {
		fmt.Fprintf(os.Stderr, "sigrok-mcp-server: %v\n", err)
		os.Exit(1)
	}
}
