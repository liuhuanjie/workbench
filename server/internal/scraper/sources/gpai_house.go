package sources

// gpai_house.go 公拍网-住宅法拍（上海/北京）抓取器
//
// 公拍网（gpai.net）是司法拍卖权威辅助平台，上海地区标的最全，服务端渲染，可直接解析。
// 入口：https://s.gpai.net/sf/Search.do?q=<城市>&page=<n>
// 注意：首次访问会下发挑战 Cookie（307 + Set-Cookie），由 httpclient 的 Cookie Jar 自动处理。
//
// 替代说明：淘宝司法拍卖（sf.taobao.com）对云主机 IP 段直接封禁（cloud_ip_bl），
// 服务器无法抓取，故沪京住宅法拍改用公拍网作为阿里侧的替代数据源。

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"workbench/server/internal/scraper"
	"workbench/server/internal/store"
)

const (
	gpaiSearchFmt = "https://s.gpai.net/sf/Search.do?q=%s&page=%d"
	gpaiItemBase  = "https://www.gpai.net/sf/item2.do?Web_Item_ID="
	gpaiMaxPages  = 3 // 单城市最多翻 3 页（60 条）
)

var gpaiCities = []string{"上海", "北京"}

var gpaiTimeRe = regexp.MustCompile(`(\d{4})-(\d{1,2})-(\d{1,2})[^\d]{0,3}(\d{1,2}):(\d{2})`)

type gpaiHouseScraper struct{}

func (gpaiHouseScraper) Key() string      { return "gpai_house" }
func (gpaiHouseScraper) Name() string     { return "公拍网-住宅法拍(沪/京)" }
func (gpaiHouseScraper) Category() string { return "house" }

func (gpaiHouseScraper) Fetch(ctx context.Context) ([]store.Item, error) {
	items := []store.Item{}
	seen := map[string]bool{}
	for _, city := range gpaiCities {
		for page := 1; page <= gpaiMaxPages; page++ {
			api := fmt.Sprintf(gpaiSearchFmt, url.QueryEscape(city), page)
			body, err := scraper.Get(ctx, api)
			if err != nil {
				if len(items) > 0 {
					break
				}
				return nil, err
			}
			pageItems, err := parseGPaiList(body, city)
			if err != nil {
				if len(items) > 0 {
					break
				}
				return nil, err
			}
			if len(pageItems) == 0 {
				break
			}
			for _, it := range pageItems {
				if seen[it.URL] {
					continue
				}
				seen[it.URL] = true
				items = append(items, it)
			}
		}
	}
	return items, nil
}

func parseGPaiList(body []byte, city string) ([]store.Item, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	items := []store.Item{}
	doc.Find("div.list-item").Each(func(_ int, sel *goquery.Selection) {
		a := sel.Find(".item-tit a").First()
		title := strings.TrimSpace(a.Text())
		href, ok := a.Attr("href")
		if title == "" || !ok {
			return
		}
		id := gpaiItemID(href)
		if id == "" {
			return
		}
		if !isResidential(title) {
			return
		}
		item := store.Item{
			SourceKey: "gpai_house",
			URL:       gpaiItemBase + id,
			Title:     title,
			City:      city,
			District:  gpaiDistrict(title),
			Status:    strings.TrimSpace(sel.Find(".badge-icon, .status-badge").First().Text()),
		}
		// 起拍价：首个 price-red（单位元）→ 万元
		if v := parseAmount(sel.Find("b.price-red").First().Text()); v > 0 {
			item.StartPriceWan = round2(v / 10000)
		}
		// 评估价/市场价：可能以万元或直接以元给出
		sel.Find("p").Each(func(_ int, p *goquery.Selection) {
			t := p.Text()
			if !strings.Contains(t, "评估价") && !strings.Contains(t, "市场价") {
				return
			}
			v := parseAmount(t)
			if v == 0 {
				return
			}
			if !strings.Contains(t, "万元") {
				v = v / 10000
			}
			item.EvalPriceWan = round2(v)
		})
		if m := gpaiTimeRe.FindStringSubmatch(sel.Text()); m != nil {
			item.AuctionDate = normalizeDate(fmt.Sprintf("%s-%s-%s %s:%s",
				m[1], m[2], m[3], m[4], m[5]))
		}
		items = append(items, item)
	})
	return items, nil
}

// gpaiItemID 从详情页链接提取标的内码
func gpaiItemID(href string) string {
	i := strings.Index(href, "Web_Item_ID=")
	if i < 0 {
		return ""
	}
	id := href[i+len("Web_Item_ID="):]
	if j := strings.IndexAny(id, "&?#"); j >= 0 {
		id = id[:j]
	}
	return strings.TrimSpace(id)
}

// gpaiDistrict 从标题提取区县，如"上海市宝山区..." → 宝山区
func gpaiDistrict(title string) string {
	for _, city := range []string{"上海市", "北京市", "上海", "北京"} {
		if i := strings.Index(title, city); i >= 0 {
			rest := title[i+len(city):]
			for _, suffix := range []string{"区", "县", "市"} {
				if j := strings.Index(rest, suffix); j >= 0 && j <= 6 {
					return rest[:j+len(suffix)]
				}
			}
			return city
		}
	}
	return ""
}

func round2(v float64) float64 { return float64(int64(v*100+0.5)) / 100 }

func init() { scraper.Register(gpaiHouseScraper{}) }
