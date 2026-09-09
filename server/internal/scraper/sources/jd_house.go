package sources

// jd_house.go 京东司法拍卖-住宅（上海/北京）抓取器
//
// 数据入口（实测可用，无需登录）：
//   https://paimai.jd.com/json/noticeJson?tab=0&publishSource=7&page=N
// 返回 JSONP：null([{id,title,publishTime,auctionTime,synopsis,orgName}, ...])，20 条/页
//
// 说明：京东司法拍卖标的列表（auction.jd.com / sifa.jd.com）强制跳登录，
// 因此改抓公开的"司法拍卖公告流"——标题含法院与标的物信息，详情页可跳转原公告。
// 公告本身不含起拍价，价格字段留空，以详情页为准。

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"workbench/server/internal/scraper"
	"workbench/server/internal/store"
)

const (
	jdNoticeAPIFmt = "https://paimai.jd.com/json/noticeJson?tab=0&publishSource=7&page=%d"
	jdNoticeFmt    = "https://paimai.jd.com/notice/%d"
	jdMaxPages     = 25 // 单次最多翻 25 页（500 条），控制请求量
)

// jdCities 关注城市（京东沪京标的占比低，故翻页略多于其他源）
var jdCities = []string{"上海", "北京"}

type jdNotice struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	PublishTime string `json:"publishTime"`
	AuctionTime string `json:"auctionTime"`
	Synopsis    string `json:"synopsis"`
	OrgName     string `json:"orgName"`
}

type jdHouseScraper struct{}

func (jdHouseScraper) Key() string      { return "jd_house" }
func (jdHouseScraper) Name() string     { return "京东法拍-住宅(沪/京)" }
func (jdHouseScraper) Category() string { return "house" }

// jdTargetCity 返回标题命中的目标城市（非沪京返回空）
func jdTargetCity(title string) string {
	for _, c := range jdCities {
		if cityMatch(title, c) {
			return c
		}
	}
	return ""
}

// isResidential 判断是否为住宅类标的（排除车辆、设备等）
func isResidential(title string) bool {
	for _, kw := range []string{"汽车", "车辆", "车牌", "机器设备", "股权", "船舶", "林木", "红木", "家具", "手机", "手表", "酒"} {
		if strings.Contains(title, kw) {
			return false
		}
	}
	for _, kw := range []string{"住宅", "房屋", "房产", "住房", "公寓", "商品房", "室", "宅"} {
		if strings.Contains(title, kw) {
			return true
		}
	}
	return false
}

func (jdHouseScraper) Fetch(ctx context.Context) ([]store.Item, error) {
	items := []store.Item{}
	seen := map[int64]bool{}
	for page := 1; page <= jdMaxPages; page++ {
		body, err := scraper.Get(ctx, fmt.Sprintf(jdNoticeAPIFmt, page))
		if err != nil {
			// 单页失败不中断：已抓到的部分照常入库，仅记录
			if len(items) > 0 {
				break
			}
			return nil, err
		}
		list, err := parseJDNotice(body)
		if err != nil {
			if len(items) > 0 {
				break
			}
			return nil, err
		}
		if len(list) == 0 {
			break
		}
		newInPage := 0
		for _, n := range list {
			if seen[n.ID] {
				continue
			}
			seen[n.ID] = true
			newInPage++
			city := jdTargetCity(n.Title)
			if city == "" || !isResidential(n.Title) {
				continue
			}
			items = append(items, store.Item{
				SourceKey:   "jd_house",
				URL:         fmt.Sprintf(jdNoticeFmt, n.ID),
				Title:       n.Title,
				Summary:     truncate(n.Synopsis, 200),
				City:        city,
				Court:       n.OrgName,
				AuctionDate: normalizeDate(n.AuctionTime),
				Status:      "公告",
				PublishedAt: normalizeDate(n.PublishTime),
			})
		}
		// 深分页会出现内容重复，发现本页全是已见条目即停止
		if newInPage == 0 {
			break
		}
	}
	return items, nil
}

// parseJDNotice 剥离 JSONP 外壳（null([...])）并解析
func parseJDNotice(body []byte) ([]jdNotice, error) {
	s := string(body)
	start := strings.Index(s, "(")
	end := strings.LastIndex(s, ")")
	if start < 0 || end <= start {
		return nil, errors.New("京东公告返回非 JSONP 格式（可能触发风控）")
	}
	var list []jdNotice
	if err := json.Unmarshal([]byte(s[start+1:end]), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func init() { scraper.Register(jdHouseScraper{}) }
