package scraper

// httpclient.go 统一 HTTP 客户端：真实 UA + Cookie 会话 + 按域名限速 + 退避重试
//
// 反爬规避策略（个人低频使用场景，宁可慢也不要被封 IP）：
//  1. 同一域名两次请求之间强制最小间隔（含随机抖动，避免固定节拍被识别为机器）
//  2. 失败按指数退避重试，最多 2 次；短时间内反复失败的源由 scraper.go 自动停用
//  3. 全程单数请求量：由各数据源自行限制翻页页数

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"

// minHostInterval 同一域名的最小请求间隔
const minHostInterval = 3 * time.Second

var (
	jar, _ = cookiejar.New(nil)
	// 带 Cookie 会话：部分站点（如公拍网）先下发挑战 Cookie 再放行，靠 Jar 自动携带
	httpClient = &http.Client{Timeout: 15 * time.Second, Jar: jar}
)

var (
	rateMu  sync.Mutex
	lastHit = map[string]time.Time{}
)

// waitTurn 按域名串行限速：距上次请求不足最小间隔时等待差额 + 随机抖动
func waitTurn(host string, ctx context.Context) error {
	rateMu.Lock()
	wait := time.Duration(0)
	if last, ok := lastHit[host]; ok {
		if d := minHostInterval - time.Since(last); d > 0 {
			wait = d
		}
	}
	// 记录本次请求时间（含抖动占用），让后续请求继续排队
	jitter := time.Duration(rand.Intn(1500)) * time.Millisecond
	lastHit[host] = time.Now().Add(wait + jitter)
	rateMu.Unlock()

	if wait < 0 {
		wait = 0
	}
	wait += jitter
	select {
	case <-time.After(wait):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Get GET 请求，按域名限速；失败按 3s / 9s 退避重试，最多两次
func Get(ctx context.Context, rawURL string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(3*(1<<(attempt-1))) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		if err := waitTurn(u.Host, ctx); err != nil {
			return nil, err
		}
		body, err := doGet(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func doGet(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml,application/json;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return body, nil
}
