package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dd_screen_go/internal/browser"
	"dd_screen_go/internal/config"
	"dd_screen_go/internal/httpapi"
	"dd_screen_go/internal/platform"
	"dd_screen_go/internal/render"
	"dd_screen_go/internal/storage"
	"dd_screen_go/internal/util"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	util.GlobalLogLevel = cfg.LogLevel

	if err := storage.EnsureRuntimeDirs(cfg.BaseDir); err != nil {
		fmt.Fprintf(os.Stderr, "运行时目录创建失败: %v\n", err)
		os.Exit(1)
	}

	store := storage.New(cfg.BaseDir)
	br := browser.New(cfg.ChromePath, cfg.Headless)
	platformSvc := platform.NewService(store)
	renderer := render.New(br, store)

	api := httpapi.New(httpapi.Dependencies{
		Config:   cfg,
		Store:    store,
		Browser:  br,
		Platform: platformSvc,
		Render:   renderer,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go store.StartCleaner(ctx, 30*time.Minute, 24*time.Hour)

	server := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      api.Handler(),
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		util.Log("INF", "Main", "DDScreenGO 服务启动中... 监听地址: %s", cfg.ListenAddr)
		util.Log("INF", "Main", "Swagger UI: http://%s/swagger/index.html", cfg.ListenAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			util.Log("ERR", "Main", "服务异常退出: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	util.Log("INF", "Main", "正在关闭服务...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		util.Log("ERR", "Main", "服务关闭异常: %v", err)
	}

	br.Close()
	util.Log("INF", "Main", "服务已安全退出")
}
