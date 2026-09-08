package cronjob

// cron.go 定时任务调度（时区 Asia/Shanghai，由容器 TZ 环境变量保证）

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"workbench/server/internal/backup"
	"workbench/server/internal/scraper"
	"workbench/server/internal/store"
)

// Start 启动定时任务（阻塞前调用，内部异步执行）
func Start(st *store.Store) {
	c := cron.New(cron.WithLocation(time.Local))

	// 每日 3 批错峰全量抓取
	c.AddFunc("0 8,12,18 * * *", func() {
		log.Println("[cron] 开始全量抓取")
		scraper.RunAll(st)
	})

	// 每日 0:05 已读重置（受配置控制）
	c.AddFunc("5 0 * * *", func() {
		if st.ConfigGetStr("daily_reset.read_mark", "on") == "on" {
			if err := st.ResetReadMarks(); err != nil {
				log.Printf("[cron] 已读重置失败: %v", err)
			} else {
				log.Println("[cron] 已读标记已重置")
			}
		}
	})

	// 每日 3:00 过期数据清理
	c.AddFunc("0 3 * * *", func() {
		days := st.ConfigGetInt("retention.days", 90)
		if err := st.Cleanup(days); err != nil {
			log.Printf("[cron] 数据清理失败: %v", err)
		} else {
			log.Printf("[cron] 已清理 %d 天前的过期数据（收藏条目保留）", days)
		}
	})

	// 每日 4:00 SQLite 在线备份
	c.AddFunc("0 4 * * *", func() {
		if err := backup.Run(st); err != nil {
			log.Printf("[cron] 备份失败: %v", err)
		} else {
			log.Println("[cron] 每日备份完成")
		}
	})

	c.Start()

	// 启动补抓：当日尚无成功抓取记录时立即执行一轮（保证重启不丢当天数据）
	if !st.HasSuccessRunToday() {
		log.Println("[cron] 当日无成功抓取记录，启动补抓")
		go scraper.RunAll(st)
	}
}
