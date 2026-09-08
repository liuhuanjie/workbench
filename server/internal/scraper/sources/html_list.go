package sources

// html_list.go 通用 HTML 列表页抓取器（goquery 驱动，新闻/债权共用）
//
// 适配静态列表页：给定列表容器、标题/链接/日期选择器即可抓取。
// 动态渲染页面（JS 加载）不适用，需单独实现。

import (
	"context"
	"errors"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"workbench/server/internal/scraper"
	"workbench/server/internal/store"
)

type htmlListSource struct {
	key, name, category string
	listURL             string
	regionHint          string // 债权源的地区提示

	itemSel   string // 列表项容器选择器
	titleSel  string
	linkSel   string
	dateSel   string
	amountSel string // 可选：金额选择器（债权源）
}

func (s htmlListSource) Key() string      { return s.key }
func (s htmlListSource) Name() string     { return s.name }
func (s htmlListSource) Category() string { return s.category }

func (s htmlListSource) Fetch(ctx context.Context) ([]store.Item, error) {
	if s.listURL == "" || s.itemSel == "" {
		return nil, errors.New("数据源待校准：请填写 listURL 与 CSS 选择器（见 sources/amc_news.go / cex_debt.go）")
	}
	body, err := scraper.Get(ctx, s.listURL)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	base := strings.TrimSuffix(s.listURL, "/")
	items := []store.Item{}
	doc.Find(s.itemSel).Each(func(_ int, sel *goquery.Selection) {
		title := strings.TrimSpace(sel.Find(s.titleSel).Text())
		href, _ := sel.Find(s.linkSel).Attr("href")
		if title == "" || href == "" {
			return
		}
		if !strings.HasPrefix(href, "http") {
			href = base + "/" + strings.TrimPrefix(href, "/")
		}
		date := normalizeDate(strings.TrimSpace(sel.Find(s.dateSel).Text()))
		item := store.Item{SourceKey: s.key, URL: href, Title: title, PublishedAt: date}
		if s.category == "news" {
			item.Category, item.OrgName = classifyNews(title)
		}
		if s.category == "debt" {
			item.Region = s.regionHint
			if s.amountSel != "" {
				item.AmountWan = parseAmount(sel.Find(s.amountSel).Text())
			}
			item.Status = "在售"
		}
		items = append(items, item)
	})
	return items, nil
}
