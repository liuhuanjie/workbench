package sources

// ali_house.go 阿里资产-法拍住宅（上海/北京）抓取器（暂时停用）
//
// 目标入口: https://sf.taobao.com （阿里资产/司法拍卖-住宅）
//
// 2026-09-08 实测结论（阿里云 ECS 47.116.202.33）：
//   - sf.taobao.com 首页可访问，但所有列表路径（item_list.htm / list/*.htm）均返回
//     deny_pc.html?...|cloud_ip_bl，即淘宝对云主机 IP 段直接封禁
//   - 带 Cookie、换 UA 均无效；H5 网关 h5api.m.taobao.com 可达但需正确的 mtop 接口名与签名
//   - 因此服务器侧无法直接抓取，沪京住宅法拍改由 gpai_house（公拍网）承担
//
// 若后续要启用阿里源，可选路径：
//  1. 申请淘宝开放平台 AppKey，改走官方 API（taobao.auction.gov.auctions.get）
//  2. 使用住宅代理出口 IP，绕过云主机封禁

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
	return nil, errors.New("阿里云主机 IP 被淘宝司法拍卖封禁（cloud_ip_bl），暂不可用；沪京住宅数据已由公拍网源提供")
}

func init() { scraper.Register(aliHouseScraper{}) }
