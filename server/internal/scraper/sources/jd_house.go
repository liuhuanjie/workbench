package sources

// jd_house.go 京东法拍-住宅（上海/北京）抓取器骨架
//
// 目标入口: https://sifa.jd.com （京东司法拍卖）
//
// TODO(上线校准): 京东法拍列表相对阿里更开放，但仍需实测：
//  1. 确认列表接口/页面结构（城市与住宅类目筛选参数）
//  2. 提取标的字段后在此实现解析逻辑
// 当前返回待校准错误，源状态页会显示"fail-待校准"。

import (
	"context"
	"errors"

	"workbench/server/internal/scraper"
	"workbench/server/internal/store"
)

type jdHouseScraper struct{}

func (jdHouseScraper) Key() string      { return "jd_house" }
func (jdHouseScraper) Name() string     { return "京东法拍-住宅(沪/京)" }
func (jdHouseScraper) Category() string { return "house" }

func (jdHouseScraper) Fetch(_ context.Context) ([]store.Item, error) {
	return nil, errors.New("数据源待校准：京东法拍需上线实测接口后启用")
}

func init() { scraper.Register(jdHouseScraper{}) }
