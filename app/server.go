package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	applog "github.com/example/dd-frame/pkg/log"
)

// RunServer 启动 Gin HTTP 服务器
func RunServer(cfg *Config, router *gin.Engine) {
	httpAddr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	// 优雅关闭
	go func() {
		applog.Info("HTTP server starting", "addr", httpAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			applog.Error("HTTP server error", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	applog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		applog.Error("server forced to shutdown", "err", err)
	}
	applog.Info("server exited")
}
