package scraper

// scraper.go 抓取器插件框架：接口 / 注册表 / 执行管道

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"workbench/server/internal/store"
)

// Scraper 数据源插件接口（技术方案 4.3）
type Scraper interface {
	Key() string      // source_key，全局唯一
	Name() string     // 展示名
	Category() string // news / debt / house
	Fetch(ctx context.Context) ([]store.Item, error)
}

var (
	mu       sync.Mutex
	registry = map[string]Scraper{}
	order    []string

	// 连续失败计数：达到阈值自动停用数据源，避免反复请求触发风控
	failMu     sync.Mutex
	failCounts = map[string]int{}
)

// maxConsecutiveFail 连续失败达到该次数后自动停用数据源（需人工在数据管理页重新启用）
const maxConsecutiveFail = 3

// Register 注册数据源（各源在 init() 中调用）
func Register(s Scraper) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := registry[s.Key()]; !ok {
		order = append(order, s.Key())
	}
	registry[s.Key()] = s
}

// Lookup 按 key 获取数据源（避免与 httpclient.go 的 HTTP Get 重名）
func Lookup(key string) (Scraper, bool) {
	mu.Lock()
	defer mu.Unlock()
	s, ok := registry[key]
	return s, ok
}

// All 按注册顺序返回全部数据源
func All() []Scraper {
	mu.Lock()
	defer mu.Unlock()
	list := make([]Scraper, 0, len(order))
	for _, k := range order {
		list = append(list, registry[k])
	}
	return list
}

// RunOne 执行单个数据源：超时控制 + panic 隔离 + 去重写入 + 状态回写
// 连续失败 maxConsecutiveFail 次自动停用，防止无效重试放大被封禁风险
func RunOne(st *store.Store, s Scraper) {
	st.EnsureSource(s.Key(), s.Name(), s.Category())
	if !st.SourceEnabled(s.Key()) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[scraper] %s panic: %v", s.Key(), r)
			st.UpdateSourceRun(s.Key(), "fail", fmt.Sprintf("panic: %v", r), 0)
			noteFailure(st, s.Key())
		}
	}()

	items, err := s.Fetch(ctx)
	if err != nil {
		st.UpdateSourceRun(s.Key(), "fail", err.Error(), 0)
		log.Printf("[scraper] %s 失败: %v", s.Key(), err)
		noteFailure(st, s.Key())
		return
	}

	var inserted, updated int
	switch s.Category() {
	case "news":
		inserted, updated, err = st.UpsertNews(ctx, items)
	case "debt":
		inserted, updated, err = st.UpsertDebt(ctx, items)
	case "house":
		inserted, updated, err = st.UpsertHouse(ctx, items)
	default:
		err = fmt.Errorf("未知数据源类别: %s", s.Category())
	}
	if err != nil {
		st.UpdateSourceRun(s.Key(), "fail", err.Error(), 0)
		log.Printf("[scraper] %s 写入失败: %v", s.Key(), err)
		noteFailure(st, s.Key())
		return
	}
	resetFailure(s.Key())
	st.UpdateSourceRun(s.Key(), "success", "", inserted+updated)
	log.Printf("[scraper] %s 成功: 共 %d 条（新增 %d，更新 %d）",
		s.Key(), len(items), inserted, updated)
}

// noteFailure 累计失败次数，达到阈值自动停用数据源
func noteFailure(st *store.Store, key string) {
	failMu.Lock()
	failCounts[key]++
	n := failCounts[key]
	failMu.Unlock()
	if n < maxConsecutiveFail {
		return
	}
	if err := st.SourceToggle(context.Background(), key, false); err != nil {
		log.Printf("[scraper] %s 自动停用失败: %v", key, err)
		return
	}
	log.Printf("[scraper] %s 连续失败 %d 次，已自动停用（请在数据管理页确认后重新启用）", key, n)
}

func resetFailure(key string) {
	failMu.Lock()
	failCounts[key] = 0
	failMu.Unlock()
}

// RunAll 串行执行全部启用数据源（源间间隔 30s，降低封禁风险）
func RunAll(st *store.Store) {
	for _, s := range All() {
		RunOne(st, s)
		time.Sleep(30 * time.Second)
	}
	log.Println("[scraper] 全量抓取完成")
}
