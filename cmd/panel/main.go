package main

import (
	"log"
	"os"
	"time"

	"github.com/AliceNetworks/gost-panel/internal/api"
	"github.com/AliceNetworks/gost-panel/internal/config"
	"github.com/AliceNetworks/gost-panel/internal/model"
	"github.com/AliceNetworks/gost-panel/internal/service"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := model.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 初始化服务
	svc := service.NewService(db, cfg)

	// 启动流量历史记录定时任务
	go startTrafficRecorder(svc)

	// 启动 API 服务
	server := api.NewServer(svc, cfg)

	log.Printf("GOST Panel starting on %s", cfg.ListenAddr)
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
		os.Exit(1)
	}
}

// startTrafficRecorder 启动流量记录定时任务
func startTrafficRecorder(svc *service.Service) {
	// 每分钟记录一次流量数据
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// 立即记录一次
	svc.RecordTrafficHistory()

	for range ticker.C {
		if err := svc.RecordTrafficHistory(); err != nil {
			log.Printf("Failed to record traffic history: %v", err)
		}
	}
}
