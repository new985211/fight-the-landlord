package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/alicebob/miniredis/v2"

	"github.com/palemoky/fight-the-landlord/internal/config"
	"github.com/palemoky/fight-the-landlord/internal/logger"
	"github.com/palemoky/fight-the-landlord/internal/server"
	"github.com/palemoky/fight-the-landlord/internal/ui"
)

var version = "dev"

func main() {
	mode := flag.String("mode", "all", "运行模式: all (默认) | server | client")
	serverAddr := flag.String("server", "localhost:1780", "服务器地址 (client 模式)")
	showVersion := flag.Bool("version", false, "显示版本号并退出")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ddz %s\n", version)
		return
	}

	switch strings.ToLower(*mode) {
	case "all":
		runAll()
	case "server":
		runServer()
	case "client":
		runClient(*serverAddr)
	default:
		log.Fatalf("未知模式 %q (可选: all, server, client)", *mode)
	}
}

// runAll 启动内嵌 Redis + 服务端（后台）+ 客户端 TUI（前台）
func runAll() {
	// 1. 启动内嵌 Redis
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("启动内嵌 Redis 失败: %v", err)
	}
	defer mr.Close()
	log.Printf("📦 内嵌 Redis 已启动 (addr: %s)", mr.Addr())

	// 2. 构建配置
	cfg := config.Default()
	cfg.Redis.Addr = mr.Addr()
	cfg.Redis.Password = ""
	cfg.Redis.DB = 0
	cfg.BOT.DouZeroEnabled = false // 不依赖外部 Python 服务

	// 3. 创建并启动服务端
	server.Version = version
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("创建服务端失败: %v", err)
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Println("🎮 斗地主服务端启动中...")
		if err := srv.Start(); err != nil {
			serverErr <- err
		}
	}()

	// 等待服务端绑定端口
	time.Sleep(300 * time.Millisecond)
	select {
	case err := <-serverErr:
		log.Fatalf("服务端启动失败: %v", err)
	default:
	}

	// 4. 启动客户端 TUI
	serverURL := fmt.Sprintf("ws://127.0.0.1:%d/ws", cfg.Server.Port)
	log.Printf("🎯 连接本地服务端: %s", serverURL)

	model := ui.NewOnlineModel(serverURL)
	p := tea.NewProgram(model)

	// 转发 SIGTERM 给 TUI 以触发优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	go func() {
		<-sigCh
		p.Send(tea.Quit())
	}()

	if _, err := p.Run(); err != nil {
		log.Printf("客户端异常退出: %v", err)
	}

	signal.Stop(sigCh)
	close(sigCh)

	// 5. 优雅关闭服务端
	log.Println("🔧 正在关闭服务端...")
	srv.GracefulShutdown(cfg.Game.ShutdownTimeoutDuration())
	log.Println("👋 再见！")
}

// runServer 启动内嵌 Redis + 服务端（守护进程模式）
func runServer() {
	// 1. 启动内嵌 Redis
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("启动内嵌 Redis 失败: %v", err)
	}
	defer mr.Close()
	log.Printf("📦 内嵌 Redis 已启动 (addr: %s)", mr.Addr())

	// 2. 构建配置
	cfg := config.Default()
	cfg.Redis.Addr = mr.Addr()
	cfg.Redis.Password = ""
	cfg.Redis.DB = 0
	cfg.BOT.DouZeroEnabled = false

	// 3. 创建并启动服务端
	server.Version = version
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("创建服务端失败: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("🎮 斗地主服务端启动中...")
		if err := srv.Start(); err != nil {
			log.Fatalf("服务端启动失败: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("📢 收到关闭信号，开始优雅关闭...")
	srv.GracefulShutdown(cfg.Game.ShutdownTimeoutDuration())
}

// runClient 启动 TUI 客户端（连接外部服务器）
func runClient(serverAddr string) {
	// 初始化日志
	if err := logger.Init(); err != nil {
		log.Printf("日志初始化失败: %v", err)
	}
	defer logger.Close()

	// Panic 恢复
	defer func() {
		if r := recover(); r != nil {
			logger.LogPanic(r)
			fmt.Print("\033[2J\033[H")
			fmt.Print("\033[?25h")
			fmt.Fprintf(os.Stderr, "\n[PANIC] 客户端崩溃: %v\n\n", r)
			fmt.Fprintf(os.Stderr, "详细日志已保存到: %s\n", logger.GetLogPath())
			os.Exit(1)
		}
	}()

	// 构建 WebSocket URL
	var serverURL string
	if strings.HasPrefix(serverAddr, "ws://") || strings.HasPrefix(serverAddr, "wss://") {
		serverURL = serverAddr
	} else {
		serverURL = fmt.Sprintf("ws://%s/ws", serverAddr)
	}

	logger.LogInfo("Connecting to server: %s", serverURL)

	model := ui.NewOnlineModel(serverURL)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		logger.LogError("Client error: %v", err)
		log.Printf("客户端异常退出: %v", err)
	}
}
