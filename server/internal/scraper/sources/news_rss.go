package sources

// news_rss.go 行业媒体 RSS 聚合新闻源（功能完整，填入订阅地址即可用）
//
// TODO(上线校准): 在 rssFeeds 中填入实际可用的行业媒体 RSS/Atom 地址，
// 候选方向：不良资产行业媒体、财经媒体不良资产频道等，需逐一验证可访问性与条目质量。

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"workbench/server/internal/scraper"
	"workbench/server/internal/store"
)

var rssFeeds = []string{
	// 示例（请替换为实测可用的订阅地址）：
	// "https://example-npa-media.example.com/rss.xml",
}

type rssScraper struct{}

func (rssScraper) Key() string      { return "industry_rss" }
func (rssScraper) Name() string     { return "行业媒体RSS" }
func (rssScraper) Category() string { return "news" }

// rssDoc 兼容 RSS 2.0 与 Atom 两种格式
type rssDoc struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
		} `xml:"channel>item"`
	} `xml:"rss"`
	Entries []struct {
		Title   string `xml:"title"`
		Link    struct {
			Href string `xml:"href,attr"`
			Text string `xml:",chardata"`
		} `xml:"link"`
		Summary string `xml:"summary"`
		Updated string `xml:"updated"`
	} `xml:"entry"`
}

func (rssScraper) Fetch(ctx context.Context) ([]store.Item, error) {
	if len(rssFeeds) == 0 {
		return nil, errors.New("数据源待校准：请在 server/internal/scraper/sources/news_rss.go 的 rssFeeds 中填入实际 RSS 地址")
	}
	items := []store.Item{}
	for _, feed := range rssFeeds {
		body, err := scraper.Get(ctx, feed)
		if err != nil {
			return nil, fmt.Errorf("抓取 %s 失败: %w", feed, err)
		}
		var doc rssDoc
		if err := xml.Unmarshal(body, &doc); err != nil {
			return nil, fmt.Errorf("解析 %s 失败: %w", feed, err)
		}
		for _, it := range doc.Channel.Items {
			if it.Title == "" || it.Link == "" {
				continue
			}
			category, org := classifyNews(it.Title)
			items = append(items, store.Item{
				SourceKey:   "industry_rss",
				URL:         it.Link,
				Title:       strings.TrimSpace(it.Title),
				Summary:     trimRunes(it.Description, 200),
				Category:    category,
				OrgName:     org,
				PublishedAt: normalizeDate(it.PubDate),
			})
		}
		for _, e := range doc.Entries {
			link := e.Link.Href
			if link == "" {
				link = e.Link.Text
			}
			if e.Title == "" || link == "" {
				continue
			}
			category, org := classifyNews(e.Title)
			items = append(items, store.Item{
				SourceKey:   "industry_rss",
				URL:         link,
				Title:       strings.TrimSpace(e.Title),
				Summary:     trimRunes(e.Summary, 200),
				Category:    category,
				OrgName:     org,
				PublishedAt: normalizeDate(e.Updated),
			})
		}
	}
	return items, nil
}

// trimRunes 按字符截断（摘要上限）
func trimRunes(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func init() { scraper.Register(rssScraper{}) }
