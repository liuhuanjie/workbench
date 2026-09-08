package sources

// amc_news.go 5大AMC 官网新闻源（基于通用 HTML 列表抓取器）
//
// TODO(上线校准): 逐站访问官网新闻中心页面，填写 listURL 与 CSS 选择器。
// 官网一般为静态列表页，适配成功率较高；若为动态渲染则需单独实现抓取器。

import "workbench/server/internal/scraper"

var amcNewsSources = []htmlListSource{
	{key: "cinda_news", name: "中国信达-官网新闻", category: "news",
		// listURL: "https://www.cinda.com.cn/...", itemSel: "...", ...
	},
	{key: "citic_amc_news", name: "中信金融资产-官网新闻", category: "news"},
	{key: "orient_news", name: "中国东方-官网新闻", category: "news"},
	{key: "greatwall_news", name: "中国长城-官网新闻", category: "news"},
	{key: "galaxy_news", name: "银河资产-官网新闻", category: "news"},
}

func init() {
	for _, s := range amcNewsSources {
		scraper.Register(s)
	}
}
