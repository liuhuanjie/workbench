package sources

// cex_debt.go 产交所债权挂牌源（基于通用 HTML 列表抓取器）
//
// TODO(上线校准): 逐所访问"不良资产/资产转让"挂牌栏目页，填写 listURL 与选择器。
// 候选产交所：北京产权交易所、上海联合产权交易所、广东联合产权交易中心等。

import "workbench/server/internal/scraper"

var cexDebtSources = []htmlListSource{
	{key: "cex_bj", name: "北交所-债权挂牌", category: "debt", regionHint: "北京"},
	{key: "cex_sh", name: "上海联交所-债权挂牌", category: "debt", regionHint: "上海"},
	{key: "cex_gd", name: "广东联合-债权挂牌", category: "debt", regionHint: "广东"},
}

func init() {
	for _, s := range cexDebtSources {
		scraper.Register(s)
	}
}
