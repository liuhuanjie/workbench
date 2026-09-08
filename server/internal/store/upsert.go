package store

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// upsert.go 三层去重写入（技术方案 3.8）：
//  1. URL 规范化 + 唯一键（主去重）
//  2. 内容指纹（跨平台/重挂去重，debt/house）
//  3. 字段级刷新（终态条目停止更新，最后快照留库）

// 追踪参数黑名单（剥离）
var trackingParams = map[string]bool{
	"utm_source": true, "utm_medium": true, "utm_campaign": true, "utm_term": true,
	"utm_content": true, "spm": true, "spm_id": true, "from": true, "scm": true,
	"referrer": true, "track_id": true, "jht": true,
}

// NormalizeURL 剥离追踪参数与 fragment，返回规范化 URL
func NormalizeURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return strings.TrimSpace(raw)
	}
	q := u.Query()
	for k := range q {
		if trackingParams[strings.ToLower(k)] {
			q.Del(k)
		}
	}
	u.RawQuery = q.Encode()
	u.Fragment = ""
	u.Host = strings.ToLower(u.Host)
	return u.String()
}

// fingerprint 内容指纹：sha1(各部分 join)
func fingerprint(parts ...string) string {
	h := sha1.New()
	for _, p := range parts {
		h.Write([]byte(strings.TrimSpace(p)))
		h.Write([]byte{'|'})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func normTitle(t string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(t)), "")
}

func now() string { return time.Now().Format("2006-01-02 15:04:05") }

// 终态集合：进入终态后停止刷新
func isFinalStatus(st string) bool {
	switch st {
	case "已成交", "已结束", "流拍", "成交", "结束":
		return true
	}
	return false
}

// UpsertNews 新闻写入（URL 主去重，指纹仅辅助更新定位）
func (s *Store) UpsertNews(ctx context.Context, items []Item) (inserted, updated int, err error) {
	for _, it := range items {
		if it.Title == "" || it.URL == "" {
			continue
		}
		u := NormalizeURL(it.URL)
		fp := fingerprint("news", normTitle(it.Title))
		var id int64
		// 指纹命中 → 同文不同链接，更新原记录
		if err := s.DB.QueryRowContext(ctx,
			`SELECT id FROM news WHERE fingerprint=?`, fp).Scan(&id); err == nil {
			_, err := s.DB.ExecContext(ctx,
				`UPDATE news SET url=?, summary=?, fetched_at=? WHERE id=?`,
				u, it.Summary, now(), id)
			if err != nil {
				return inserted, updated, err
			}
			updated++
			continue
		}
		// URL 命中 → 刷新动态字段
		if err := s.DB.QueryRowContext(ctx,
			`SELECT id FROM news WHERE url=?`, u).Scan(&id); err == nil {
			_, err := s.DB.ExecContext(ctx,
				`UPDATE news SET summary=?, fetched_at=? WHERE id=?`,
				it.Summary, now(), id)
			if err != nil {
				return inserted, updated, err
			}
			updated++
			continue
		}
		if _, err := s.DB.ExecContext(ctx,
			`INSERT INTO news(source_key, url, title, summary, category, org_name, published_at, fetched_at, fingerprint)
			 VALUES(?,?,?,?,?,?,?,?,?)`,
			it.SourceKey, u, it.Title, it.Summary, defaultStr(it.Category, "industry"),
			it.OrgName, it.PublishedAt, now(), fp); err != nil {
			return inserted, updated, err
		}
		inserted++
	}
	return inserted, updated, nil
}

// UpsertDebt 债权写入（URL + 内容指纹双重去重）
func (s *Store) UpsertDebt(ctx context.Context, items []Item) (inserted, updated int, err error) {
	for _, it := range items {
		if it.Title == "" || it.URL == "" {
			continue
		}
		u := NormalizeURL(it.URL)
		fp := fingerprint("debt", normTitle(it.Title), it.Region,
			fmt.Sprintf("%.0f", it.AmountWan), it.EndTime)
		var id int64
		var curStatus string
		// 指纹命中 → 同标的跨平台/重挂，更新原记录
		if err := s.DB.QueryRowContext(ctx,
			`SELECT id FROM debt_item WHERE fingerprint=?`, fp).Scan(&id); err == nil {
			_ = s.DB.QueryRowContext(ctx,
				`SELECT status FROM debt_item WHERE id=?`, id).Scan(&curStatus)
			if !isFinalStatus(curStatus) {
				_, err := s.DB.ExecContext(ctx,
					`UPDATE debt_item SET url=?, status=?, end_time=?, amount_wan=?, fetched_at=? WHERE id=?`,
					u, it.Status, it.EndTime, it.AmountWan, now(), id)
				if err != nil {
					return inserted, updated, err
				}
			}
			updated++
			continue
		}
		// URL 命中 → 字段级刷新
		if err := s.DB.QueryRowContext(ctx,
			`SELECT id, status FROM debt_item WHERE url=?`, u).Scan(&id, &curStatus); err == nil {
			if !isFinalStatus(curStatus) {
				_, err := s.DB.ExecContext(ctx,
					`UPDATE debt_item SET status=?, end_time=?, amount_wan=?, fetched_at=? WHERE id=?`,
					it.Status, it.EndTime, it.AmountWan, now(), id)
				if err != nil {
					return inserted, updated, err
				}
			}
			updated++
			continue
		}
		if _, err := s.DB.ExecContext(ctx,
			`INSERT INTO debt_item(source_key, url, title, summary, transferor, amount_wan, region, status, end_time, published_at, fetched_at, fingerprint)
			 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
			it.SourceKey, u, it.Title, it.Summary, it.Transferor, it.AmountWan,
			it.Region, it.Status, it.EndTime, it.PublishedAt, now(), fp); err != nil {
			return inserted, updated, err
		}
		inserted++
	}
	return inserted, updated, nil
}

// UpsertHouse 住宅法拍写入（URL + 内容指纹双重去重）
func (s *Store) UpsertHouse(ctx context.Context, items []Item) (inserted, updated int, err error) {
	for _, it := range items {
		if it.Title == "" || it.URL == "" {
			continue
		}
		u := NormalizeURL(it.URL)
		fp := fingerprint("house", normTitle(it.Title), it.City, it.District,
			fmt.Sprintf("%.0f", it.StartPriceWan), it.AuctionDate)
		var id int64
		var curStatus string
		if err := s.DB.QueryRowContext(ctx,
			`SELECT id FROM house_item WHERE fingerprint=?`, fp).Scan(&id); err == nil {
			_ = s.DB.QueryRowContext(ctx,
				`SELECT status FROM house_item WHERE id=?`, id).Scan(&curStatus)
			if !isFinalStatus(curStatus) {
				_, err := s.DB.ExecContext(ctx,
					`UPDATE house_item SET url=?, status=?, start_price_wan=?, eval_price_wan=?, auction_date=?, fetched_at=? WHERE id=?`,
					u, it.Status, it.StartPriceWan, it.EvalPriceWan, it.AuctionDate, now(), id)
				if err != nil {
					return inserted, updated, err
				}
			}
			updated++
			continue
		}
		if err := s.DB.QueryRowContext(ctx,
			`SELECT id, status FROM house_item WHERE url=?`, u).Scan(&id, &curStatus); err == nil {
			if !isFinalStatus(curStatus) {
				_, err := s.DB.ExecContext(ctx,
					`UPDATE house_item SET status=?, start_price_wan=?, eval_price_wan=?, auction_date=?, fetched_at=? WHERE id=?`,
					it.Status, it.StartPriceWan, it.EvalPriceWan, it.AuctionDate, now(), id)
				if err != nil {
					return inserted, updated, err
				}
			}
			updated++
			continue
		}
		if _, err := s.DB.ExecContext(ctx,
			`INSERT INTO house_item(source_key, url, title, city, district, area_sqm, start_price_wan, eval_price_wan, court, auction_date, status, published_at, fetched_at, fingerprint)
			 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			it.SourceKey, u, it.Title, it.City, it.District, it.AreaSqm,
			it.StartPriceWan, it.EvalPriceWan, it.Court, it.AuctionDate,
			it.Status, it.PublishedAt, now(), fp); err != nil {
			return inserted, updated, err
		}
		inserted++
	}
	return inserted, updated, nil
}

func defaultStr(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
