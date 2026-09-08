package store

// Item 抓取结果统一模型（新闻/债权/法拍共用，不适用的字段留空）
type Item struct {
	SourceKey string
	URL       string
	Title     string
	Summary   string
	Category  string // news 分类：central_amc / local_amc / bank_transfer / industry
	OrgName   string

	// 债权/法拍专属
	Transferor    string
	AmountWan     float64
	Region        string
	Status        string
	EndTime       string
	City          string
	District      string
	AreaSqm       float64
	StartPriceWan float64
	EvalPriceWan  float64
	Court         string
	AuctionDate   string

	PublishedAt string
}
