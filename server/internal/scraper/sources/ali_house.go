package sources

// ali_house.go 阿里资产-法拍住宅（上海/北京）抓取器骨架
//
// 目标入口: https://sf.taobao.com （阿里资产/司法拍卖-住宅）
//
// TODO(上线校准): 阿里资产列表为动态接口且带反爬（cookie/加密参数），
// 需上线实测接口地址、参数与所需请求头后实现本抓取器：
//  1. 浏览器 F12 抓取列表请求（JSON 接口）
//  2. 提取 itemId、标题、区域、面积、起拍价、评估价、开拍时间、状态
//  3. 过滤城市（上海/北京）与住宅类目
//  4. 用 scraper.Get 携带必要头请求接口并解析
// 当前返回待校准错误，源状态页会显示"fail-待校准"。

import (
	"context"
	"errors"

	"workbench/server/internal/scraper"
	"workbench/server/internal/store"
)

type aliHouseScraper struct{}

func (aliHouseScraper) Key() string      { return "ali_house" }
func (aliHouseScraper) Name() string     { return "阿里资产-法拍住宅(沪/京)" }
func (aliHouseScraper) Category() string { return "house" }

func (aliHouseScraper) Fetch(_ context.Context) ([]store.Item, error) {
	return nil, errors.New("数据源待校准：阿里资产需上线实测接口与反爬参数后启用")
}

func init() { scraper.Register(aliHouseScraper{}) }
