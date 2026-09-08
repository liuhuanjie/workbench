package scraper

// httpclient.go 统一 HTTP 客户端：真实 UA + 超时 + 单次重试

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Get GET 请求，失败退避 3 秒后重试一次
func Get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(3 * time.Second):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml,application/json;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}
		return body, nil
	}
	return nil, lastErr
}
