package main

// main.go 入口：装配 store / auth / scraper / cron / api

import (
	"log"
	"net/http"
	"time"

	"workbench/server/internal/api"
	"workbench/server/internal/auth"
	"workbench/server/internal/config"
	"workbench/server/internal/cronjob"
	"workbench/server/internal/scraper"
	"workbench/server/internal/store"

	// 注册全部数据源（init 副作用）
	_ "workbench/server/internal/scraper/sources"
)

func main() {
	cfg := config.Load()

	st, err := store.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer st.Close()

	// 首次启动初始化 admin（随机密码打印到日志）
	authSvc := auth.NewService(st, cfg.JWTSecret)
	authSvc.EnsureAdmin()

	// 注册全部数据源状态行（数据管理页可见）
	for _, s := range scraper.All() {
		st.EnsureSource(s.Key(), s.Name(), s.Category())
	}

	// 定时任务（含启动补抓）
	go cronjob.Start(st)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewRouter(st, authSvc),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("server listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("服务退出: %v", err)
	}
}
