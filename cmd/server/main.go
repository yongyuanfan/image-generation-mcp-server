package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/yong/image-generation-mcp-server/internal/apiserver"
	"github.com/yong/image-generation-mcp-server/internal/config"
	"github.com/yong/image-generation-mcp-server/internal/mcpserver"
	"github.com/yong/image-generation-mcp-server/internal/provider/ark"
	imagesvc "github.com/yong/image-generation-mcp-server/internal/service/image"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	provider := ark.NewClient(cfg)
	service := imagesvc.NewService(cfg, provider)

	mcpHandler := mcpserver.NewHandler(cfg, service)
	apiHandler := apiserver.NewHandler(cfg, service)

	mux := http.NewServeMux()
	mux.Handle(cfg.MCPPath, mcpHandler)
	mux.Handle(cfg.MCPPath+"/", mcpHandler)
	mux.Handle(cfg.APIPrefix+"/", http.StripPrefix(cfg.APIPrefix, apiHandler))
	mux.HandleFunc("/healthz", apiserver.Healthz)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("http server listening on %s", cfg.HTTPAddr)
		log.Printf("mcp endpoint available at %s", cfg.MCPPath)
		log.Printf("api endpoint available at %s", cfg.APIPrefix)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
	log.Print("server stopped")
}
