package main

import (
	"fmt"
	"os"

	"github.com/KenosInc/sigrok-mcp-server/internal/config"
	"github.com/KenosInc/sigrok-mcp-server/internal/devices"
	"github.com/KenosInc/sigrok-mcp-server/internal/serial"
	"github.com/KenosInc/sigrok-mcp-server/internal/sigrok"
	"github.com/KenosInc/sigrok-mcp-server/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	cfg := config.Load()
	executor := sigrok.NewExecutor(cfg.SigrokCLIPath, cfg.Timeout, cfg.WorkingDir)

	deviceRegistry, err := devices.LoadEmbedded()
	if err != nil {
		fmt.Fprintf(os.Stderr, "sigrok-mcp-server: load device profiles: %v\n", err)
		os.Exit(1)
	}

	handlers := tools.NewHandlers(executor, config.FirmwareDirs(), serial.NewPortQuerier(), deviceRegistry)

	srv := server.NewMCPServer("sigrok-mcp-server", "0.1.0",
		server.WithResourceCapabilities(false, false),
	)
	tools.RegisterAll(srv, handlers)
	tools.RegisterResources(srv, deviceRegistry)

	if err := server.ServeStdio(srv); err != nil {
		fmt.Fprintf(os.Stderr, "sigrok-mcp-server: %v\n", err)
		os.Exit(1)
	}
}
