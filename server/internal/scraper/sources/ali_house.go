package sources

// ali_house.go 阿里司法拍卖-住宅（淘宝开放平台 官方 API）
//
// 背景：直接抓取 sf.taobao.com 会因云主机 IP 被封禁（cloud_ip_bl）而不可用，
// 因此改为调用淘宝开放平台的官方司法拍卖 API：
//
//	taobao.auction.gov.auctions.get  分页获取标的物信息（免费、不需要用户授权）
//
// 官方接口清单（developer.alibaba.com 文档 apiId=25290）：
//	cat_ids / court_city / court_name / court_prov / item_city / item_prov
//	current_page / page_size / sort(0价格升,1价格降,2出价数降) / status
// 返回：auctions[] 含 item_id/title/court_name/auction_status/starting_price/
//	consult_price/current_price/start_time/end_time 等
//
// 启用条件：配置环境变量 TAOBAO_APP_KEY 与 TAOBAO_APP_SECRET（见 .env.example），
// 未配置时本源不注册（数据管理页不出现），避免无效失败。

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"workbench/server/internal/scraper"
	"workbench/server/internal/store"
)

const (
	taobaoGateway = "https://eco.taobao.com/router/rest"
	taobaoItemFmt = "https://sf.taobao.com/sf_item/%d.htm"
	taobaoMethod  = "taobao.auction.gov.auctions.get"
	taobaoMaxPage = 3 // 单城市最多翻 3 页
)

// taobaoCities 关注城市（标的物所在城市，不带"市"字，与官方示例"杭州"一致）
var taobaoCities = []string{"上海", "北京"}

var taobaoClient = &http.Client{Timeout: 15 * time.Second}

type taobaoAuction struct {
	ItemID        int64  `json:"item_id"`
	Title         string `json:"title"`
	CourtName     string `json:"court_name"`
	Status        string `json:"auction_status"`
	StartingPrice string `json:"starting_price"`
	ConsultPrice  string `json:"consult_price"`
	CurrentPrice  string `json:"current_price"`
	StartTime     int64  `json:"start_time"`
	EndTime       int64  `json:"end_time"`
}

type taobaoResponse struct {
	Resp struct {
		Auctions json.RawMessage `json:"auctions"`
		Total    int64           `json:"total"`
	} `json:"auction_gov_auctions_get_response"`
	Error *struct {
		Code    string `json:"code"`
		Msg     string `json:"msg"`
		SubCode string `json:"sub_code"`
		SubMsg  string `json:"sub_msg"`
	} `json:"error_response"`
}

type aliHouseScraper struct {
	appKey    string
	appSecret string
}

func (aliHouseScraper) Key() string      { return "ali_house" }
func (aliHouseScraper) Name() string     { return "阿里法拍-住宅(官方API)" }
func (aliHouseScraper) Category() string { return "house" }

// taobaoSign TOP 签名：secret + 参数按 key 升序拼接 key+value + secret，MD5 后转大写
func taobaoSign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(secret)
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(params[k])
	}
	b.WriteString(secret)
	sum := md5.Sum([]byte(b.String()))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func (s aliHouseScraper) call(ctx context.Context, city string, page int) ([]taobaoAuction, error) {
	params := map[string]string{
		"method":       taobaoMethod,
		"app_key":      s.appKey,
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
		"format":       "json",
		"v":            "2.0",
		"sign_method":  "md5",
		"item_city":    city,
		"current_page": strconv.Itoa(page),
		"page_size":    "20",
		"sort":         "0",
	}
	params["sign"] = taobaoSign(params, s.appSecret)

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, taobaoGateway, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := taobaoClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out taobaoResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	if out.Error != nil {
		msg := out.Error.Msg
		if out.Error.SubMsg != "" {
			msg += " / " + out.Error.SubMsg
		}
		return nil, fmt.Errorf("淘宝API错误(%s): %s", out.Error.Code, msg)
	}
	raw := out.Resp.Auctions
	if len(raw) == 0 {
		return nil, nil
	}
	// TOP 协议里数组字段可能直接是数组，也可能包一层 {"auction":[...]}
	var list []taobaoAuction
	if err := json.Unmarshal(raw, &list); err != nil {
		var wrap struct {
			Auction []taobaoAuction `json:"auction"`
		}
		if err2 := json.Unmarshal(raw, &wrap); err2 != nil {
			return nil, err
		}
		list = wrap.Auction
	}
	return list, nil
}

func (s aliHouseScraper) Fetch(ctx context.Context) ([]store.Item, error) {
	items := []store.Item{}
	seen := map[int64]bool{}
	for _, city := range taobaoCities {
		for page := 1; page <= taobaoMaxPage; page++ {
			list, err := s.call(ctx, city, page)
			if err != nil {
				return nil, err
			}
			if len(list) == 0 {
				break
			}
			for _, a := range list {
				if seen[a.ItemID] || a.ItemID == 0 {
					continue
				}
				seen[a.ItemID] = true
				items = append(items, store.Item{
					SourceKey:     "ali_house",
					URL:           fmt.Sprintf(taobaoItemFmt, a.ItemID),
					Title:         a.Title,
					City:          city,
					District:      gpaiDistrict(a.Title),
					StartPriceWan: yuanToWan(a.StartingPrice),
					EvalPriceWan:  yuanToWan(a.ConsultPrice),
					Court:         a.CourtName,
					AuctionDate:   msTime(a.StartTime),
					EndTime:       msTime(a.EndTime),
					Status:        a.Status,
					Summary:       fmt.Sprintf("当前价 %s 元", a.CurrentPrice),
				})
			}
		}
	}
	return items, nil
}

// yuanToWan 官方接口金额单位为元，统一换算为万元（解析失败返回 0）
func yuanToWan(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || v <= 0 {
		return 0
	}
	return round2(v / 10000)
}

func msTime(ms int64) string {
	if ms <= 0 {
		return ""
	}
	return time.UnixMilli(ms).Format("2006-01-02 15:04")
}

// 仅在配置了淘宝开放平台密钥时注册本数据源
func init() {
	key := os.Getenv("TAOBAO_APP_KEY")
	secret := os.Getenv("TAOBAO_APP_SECRET")
	if key == "" || secret == "" {
		return
	}
	scraper.Register(aliHouseScraper{appKey: key, appSecret: secret})
}
