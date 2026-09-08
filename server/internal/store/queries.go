package store

// queries.go 全部业务查询：列表/已读/收藏/统计/配置/数据源状态/维护

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ---------- 通用列表 ----------

// Page 分页参数
type Page struct {
	Page     int64
	PageSize int64
}

func (p Page) norm() (int64, int64) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
	return p.Page, p.PageSize
}

func (p Page) offset() int {
	pg, size := p.norm()
	return int((pg - 1) * size)
}

func (p Page) limit() int {
	_, size := p.norm()
	return int(size)
}

// NewsFilter 新闻筛选
type NewsFilter struct {
	Category string
	Keyword  string
	Unread   bool
	Page     Page
}

type NewsRow struct {
	ID          int64  `json:"id"`
	SourceKey   string `json:"source_key"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Category    string `json:"category"`
	OrgName     string `json:"org_name"`
	PublishedAt string `json:"published_at"`
	FetchedAt   string `json:"fetched_at"`
	IsRead      int    `json:"is_read"`
	FavID       int64  `json:"fav_id"` // >0 表示已收藏
}

func like(kw string) string { return "%" + kw + "%" }

func (s *Store) ListNews(ctx context.Context, f NewsFilter) (int64, []NewsRow, error) {
	where, args := []string{"1=1"}, []any{}
	if f.Category != "" {
		where = append(where, "category = ?")
		args = append(args, f.Category)
	}
	if f.Keyword != "" {
		where = append(where, "(title LIKE ? OR org_name LIKE ?)")
		args = append(args, like(f.Keyword), like(f.Keyword))
	}
	if f.Unread {
		where = append(where, "is_read = 0")
	}
	w := strings.Join(where, " AND ")
	var total int64
	if err := s.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM news WHERE "+w, args...).Scan(&total); err != nil {
		return 0, nil, err
	}
	q := `SELECT n.id, n.source_key, n.url, n.title, n.summary, n.category, n.org_name,
	             n.published_at, n.fetched_at, n.is_read, IFNULL(f.id, 0)
	      FROM news n LEFT JOIN favorite f ON f.item_type='news' AND f.item_id = n.id
	      WHERE ` + w + ` ORDER BY n.fetched_at DESC, n.id DESC LIMIT ? OFFSET ?`
	rows, err := s.DB.QueryContext(ctx, q, append(args, f.Page.limit(), f.Page.offset())...)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	list := []NewsRow{}
	for rows.Next() {
		var r NewsRow
		if err := rows.Scan(&r.ID, &r.SourceKey, &r.URL, &r.Title, &r.Summary, &r.Category,
			&r.OrgName, &r.PublishedAt, &r.FetchedAt, &r.IsRead, &r.FavID); err != nil {
			return 0, nil, err
		}
		list = append(list, r)
	}
	return total, list, rows.Err()
}

// DebtFilter 债权筛选
type DebtFilter struct {
	SourceKey string
	Region    string
	Status    string
	Keyword   string
	Page      Page
}

type DebtRow struct {
	ID          int64   `json:"id"`
	SourceKey   string  `json:"source_key"`
	URL         string  `json:"url"`
	Title       string  `json:"title"`
	Transferor  string  `json:"transferor"`
	AmountWan   float64 `json:"amount_wan"`
	Region      string  `json:"region"`
	Status      string  `json:"status"`
	EndTime     string  `json:"end_time"`
	PublishedAt string  `json:"published_at"`
	FetchedAt   string  `json:"fetched_at"`
	IsRead      int     `json:"is_read"`
	FavID       int64   `json:"fav_id"`
}

func (s *Store) ListDebt(ctx context.Context, f DebtFilter) (int64, []DebtRow, error) {
	where, args := []string{"1=1"}, []any{}
	if f.SourceKey != "" {
		where = append(where, "source_key = ?")
		args = append(args, f.SourceKey)
	}
	if f.Region != "" {
		where = append(where, "region LIKE ?")
		args = append(args, like(f.Region))
	}
	if f.Status != "" {
		where = append(where, "status = ?")
		args = append(args, f.Status)
	}
	if f.Keyword != "" {
		where = append(where, "(title LIKE ? OR transferor LIKE ?)")
		args = append(args, like(f.Keyword), like(f.Keyword))
	}
	w := strings.Join(where, " AND ")
	var total int64
	if err := s.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM debt_item WHERE "+w, args...).Scan(&total); err != nil {
		return 0, nil, err
	}
	q := `SELECT d.id, d.source_key, d.url, d.title, d.transferor, d.amount_wan, d.region,
	             d.status, d.end_time, d.published_at, d.fetched_at, d.is_read, IFNULL(f.id, 0)
	      FROM debt_item d LEFT JOIN favorite f ON f.item_type='debt' AND f.item_id = d.id
	      WHERE ` + w + ` ORDER BY d.fetched_at DESC, d.id DESC LIMIT ? OFFSET ?`
	rows, err := s.DB.QueryContext(ctx, q, append(args, f.Page.limit(), f.Page.offset())...)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	list := []DebtRow{}
	for rows.Next() {
		var r DebtRow
		if err := rows.Scan(&r.ID, &r.SourceKey, &r.URL, &r.Title, &r.Transferor, &r.AmountWan,
			&r.Region, &r.Status, &r.EndTime, &r.PublishedAt, &r.FetchedAt, &r.IsRead, &r.FavID); err != nil {
			return 0, nil, err
		}
		list = append(list, r)
	}
	return total, list, rows.Err()
}

// HouseFilter 住宅法拍筛选
type HouseFilter struct {
	City       string
	District   string
	Status     string
	DateFrom   string
	DateTo     string
	Keyword    string
	Page       Page
}

type HouseRow struct {
	ID            int64   `json:"id"`
	SourceKey     string  `json:"source_key"`
	URL           string  `json:"url"`
	Title         string  `json:"title"`
	City          string  `json:"city"`
	District      string  `json:"district"`
	AreaSqm       float64 `json:"area_sqm"`
	StartPriceWan float64 `json:"start_price_wan"`
	EvalPriceWan  float64 `json:"eval_price_wan"`
	Court         string  `json:"court"`
	AuctionDate   string  `json:"auction_date"`
	Status        string  `json:"status"`
	FetchedAt     string  `json:"fetched_at"`
	IsRead        int     `json:"is_read"`
	FavID         int64   `json:"fav_id"`
}

func (s *Store) ListHouse(ctx context.Context, f HouseFilter) (int64, []HouseRow, error) {
	where, args := []string{"1=1"}, []any{}
	if f.City != "" {
		where = append(where, "city = ?")
		args = append(args, f.City)
	}
	if f.District != "" {
		where = append(where, "district LIKE ?")
		args = append(args, like(f.District))
	}
	if f.Status != "" {
		where = append(where, "status = ?")
		args = append(args, f.Status)
	}
	if f.DateFrom != "" {
		where = append(where, "auction_date >= ?")
		args = append(args, f.DateFrom)
	}
	if f.DateTo != "" {
		where = append(where, "auction_date <= ?")
		args = append(args, f.DateTo+" 23:59:59")
	}
	if f.Keyword != "" {
		where = append(where, "(title LIKE ? OR district LIKE ?)")
		args = append(args, like(f.Keyword), like(f.Keyword))
	}
	w := strings.Join(where, " AND ")
	var total int64
	if err := s.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM house_item WHERE "+w, args...).Scan(&total); err != nil {
		return 0, nil, err
	}
	q := `SELECT h.id, h.source_key, h.url, h.title, h.city, h.district, h.area_sqm,
	             h.start_price_wan, h.eval_price_wan, h.court, h.auction_date, h.status,
	             h.fetched_at, h.is_read, IFNULL(f.id, 0)
	      FROM house_item h LEFT JOIN favorite f ON f.item_type='house' AND f.item_id = h.id
	      WHERE ` + w + ` ORDER BY h.fetched_at DESC, h.id DESC LIMIT ? OFFSET ?`
	rows, err := s.DB.QueryContext(ctx, q, append(args, f.Page.limit(), f.Page.offset())...)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	list := []HouseRow{}
	for rows.Next() {
		var r HouseRow
		if err := rows.Scan(&r.ID, &r.SourceKey, &r.URL, &r.Title, &r.City, &r.District,
			&r.AreaSqm, &r.StartPriceWan, &r.EvalPriceWan, &r.Court, &r.AuctionDate,
			&r.Status, &r.FetchedAt, &r.IsRead, &r.FavID); err != nil {
			return 0, nil, err
		}
		list = append(list, r)
	}
	return total, list, rows.Err()
}

// ---------- 已读 / 收藏 ----------

var readTables = map[string]string{"news": "news", "debt": "debt_item", "house": "house_item"}

// MarkRead 批量标记已读
func (s *Store) MarkRead(ctx context.Context, itemType string, ids []int64) error {
	table, ok := readTables[itemType]
	if !ok || len(ids) == 0 {
		return fmt.Errorf("无效的条目类型或空列表")
	}
	ph := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		ph[i] = "?"
		args[i] = id
	}
	_, err := s.DB.ExecContext(ctx,
		"UPDATE "+table+" SET is_read = 1 WHERE id IN ("+strings.Join(ph, ",")+")", args...)
	return err
}

// ResetReadMarks 每日已读重置
func (s *Store) ResetReadMarks() error {
	for _, table := range readTables {
		if _, err := s.DB.Exec("UPDATE " + table + " SET is_read = 0"); err != nil {
			return err
		}
	}
	return nil
}

// FavoriteAdd 新增收藏
func (s *Store) FavoriteAdd(ctx context.Context, itemType string, itemID int64, note string) error {
	if _, ok := readTables[itemType]; !ok {
		return fmt.Errorf("无效的条目类型")
	}
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO favorite(item_type, item_id, note, created_at) VALUES(?,?,?,?)
		 ON CONFLICT(item_type, item_id) DO UPDATE SET note=excluded.note`,
		itemType, itemID, note, now())
	return err
}

// FavoriteRemove 取消收藏
func (s *Store) FavoriteRemove(ctx context.Context, id int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM favorite WHERE id=?`, id)
	return err
}

type FavoriteRow struct {
	ID        int64  `json:"id"`
	ItemType  string `json:"item_type"`
	ItemID    int64  `json:"item_id"`
	Note      string `json:"note"`
	CreatedAt string `json:"created_at"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Extra     string `json:"extra"` // 类型相关的补充信息
}

// ListFavorites 收藏列表（联查各表标题/链接）
func (s *Store) ListFavorites(ctx context.Context, itemType string, p Page) (int64, []FavoriteRow, error) {
	where, args := "1=1", []any{}
	if itemType != "" {
		where = "item_type = ?"
		args = append(args, itemType)
	}
	var total int64
	if err := s.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM favorite WHERE "+where, args...).Scan(&total); err != nil {
		return 0, nil, err
	}
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, item_type, item_id, note, created_at FROM favorite WHERE `+where+
			` ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`,
		append(args, p.limit(), p.offset())...)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	list := []FavoriteRow{}
	for rows.Next() {
		var r FavoriteRow
		if err := rows.Scan(&r.ID, &r.ItemType, &r.ItemID, &r.Note, &r.CreatedAt); err != nil {
			return 0, nil, err
		}
		s.fillFavoriteDetail(ctx, &r)
		list = append(list, r)
	}
	return total, list, rows.Err()
}

func (s *Store) fillFavoriteDetail(ctx context.Context, r *FavoriteRow) {
	switch r.ItemType {
	case "news":
		s.DB.QueryRowContext(ctx,
			`SELECT title, url, org_name FROM news WHERE id=?`, r.ItemID).
			Scan(&r.Title, &r.URL, &r.Extra)
	case "debt":
		var amt float64
		if s.DB.QueryRowContext(ctx,
			`SELECT title, url, amount_wan FROM debt_item WHERE id=?`, r.ItemID).
			Scan(&r.Title, &r.URL, &amt) == nil {
			r.Extra = fmt.Sprintf("%.0f 万元", amt)
		}
	case "house":
		var price float64
		if s.DB.QueryRowContext(ctx,
			`SELECT title, url, start_price_wan FROM house_item WHERE id=?`, r.ItemID).
			Scan(&r.Title, &r.URL, &price) == nil {
			r.Extra = fmt.Sprintf("起拍 %.0f 万元", price)
		}
	}
	if r.Title == "" {
		r.Title = "（原条目已清理）"
	}
}

// ---------- 统计 ----------

type Overview struct {
	NewsToday     int64           `json:"news_today"`
	DebtToday     int64           `json:"debt_today"`
	HouseToday    int64           `json:"house_today"`
	NewsUnread    int64           `json:"news_unread"`
	DebtUnread    int64           `json:"debt_unread"`
	HouseUnread   int64           `json:"house_unread"`
	FavoriteTotal int64           `json:"favorite_total"`
	Trend         []TrendPoint    `json:"trend"`
}

type TrendPoint struct {
	Date  string `json:"date"`
	News  int64  `json:"news"`
	Debt  int64  `json:"debt"`
	House int64  `json:"house"`
}

func (s *Store) StatsOverview(ctx context.Context) (*Overview, error) {
	o := &Overview{Trend: []TrendPoint{}}
	today := time.Now().Format("2006-01-02")
	s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM news WHERE fetched_at >= ?`, today).Scan(&o.NewsToday)
	s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM debt_item WHERE fetched_at >= ?`, today).Scan(&o.DebtToday)
	s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM house_item WHERE fetched_at >= ?`, today).Scan(&o.HouseToday)
	s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM news WHERE is_read = 0`).Scan(&o.NewsUnread)
	s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM debt_item WHERE is_read = 0`).Scan(&o.DebtUnread)
	s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM house_item WHERE is_read = 0`).Scan(&o.HouseUnread)
	s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM favorite`).Scan(&o.FavoriteTotal)

	// 近 7 日趋势
	dates := []string{}
	for i := 6; i >= 0; i-- {
		dates = append(dates, time.Now().AddDate(0, 0, -i).Format("2006-01-02"))
	}
	counts := map[string]*TrendPoint{}
	for _, d := range dates {
		counts[d] = &TrendPoint{Date: d[5:]} // MM-DD
	}
	scanTrend := func(table string, field func(*TrendPoint) *int64) {
		rows, err := s.DB.QueryContext(ctx,
			`SELECT substr(fetched_at,1,10) d, COUNT(*) FROM `+table+
				` WHERE fetched_at >= ? GROUP BY d`, dates[0])
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var d string
			var n int64
			if rows.Scan(&d, &n) == nil {
				if p, ok := counts[d]; ok {
					*field(p) = n
				}
			}
		}
	}
	scanTrend("news", func(p *TrendPoint) *int64 { return &p.News })
	scanTrend("debt_item", func(p *TrendPoint) *int64 { return &p.Debt })
	scanTrend("house_item", func(p *TrendPoint) *int64 { return &p.House })
	for _, d := range dates {
		o.Trend = append(o.Trend, *counts[d])
	}
	return o, nil
}

// ---------- 配置 ----------

type ConfigRow struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *Store) ConfigAll(ctx context.Context) ([]ConfigRow, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT key, value FROM app_config ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []ConfigRow{}
	for rows.Next() {
		var r ConfigRow
		if err := rows.Scan(&r.Key, &r.Value); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *Store) ConfigSet(ctx context.Context, key, value string) error {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE app_config SET value=? WHERE key=?`, value, key)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("未知的配置项: %s", key)
	}
	return nil
}

func (s *Store) ConfigGetInt(key string, def int) int {
	var v string
	if s.DB.QueryRow(`SELECT value FROM app_config WHERE key=?`, key).Scan(&v) != nil {
		return def
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return def
	}
	return n
}

func (s *Store) ConfigGetStr(key, def string) string {
	var v string
	if s.DB.QueryRow(`SELECT value FROM app_config WHERE key=?`, key).Scan(&v) != nil {
		return def
	}
	return v
}

// ---------- 数据源状态 ----------

type SourceRow struct {
	SourceKey    string `json:"source_key"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Enabled      int    `json:"enabled"`
	LastRunAt    string `json:"last_run_at"`
	LastStatus   string `json:"last_status"`
	LastError    string `json:"last_error"`
	LastItemCount int   `json:"last_item_count"`
}

// EnsureSource 注册数据源（抓取器启动时调用，幂等）
func (s *Store) EnsureSource(key, name, category string) {
	s.DB.Exec(`INSERT INTO source_status(source_key, name, category) VALUES(?,?,?)
		ON CONFLICT(source_key) DO NOTHING`, key, name, category)
}

func (s *Store) ListSources(ctx context.Context) ([]SourceRow, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT source_key, name, category, enabled, last_run_at, last_status, last_error, last_item_count
		 FROM source_status ORDER BY category, source_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []SourceRow{}
	for rows.Next() {
		var r SourceRow
		if err := rows.Scan(&r.SourceKey, &r.Name, &r.Category, &r.Enabled, &r.LastRunAt,
			&r.LastStatus, &r.LastError, &r.LastItemCount); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *Store) SourceToggle(ctx context.Context, key string, enabled bool) error {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE source_status SET enabled=? WHERE source_key=?`, boolToInt(enabled), key)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("未知数据源: %s", key)
	}
	return nil
}

func (s *Store) SourceEnabled(key string) bool {
	var enabled int
	if s.DB.QueryRow(`SELECT enabled FROM source_status WHERE source_key=?`, key).Scan(&enabled) != nil {
		return false
	}
	return enabled == 1
}

// UpdateSourceRun 回写抓取结果
func (s *Store) UpdateSourceRun(key, status, errMsg string, itemCount int) {
	if len(errMsg) > 500 {
		errMsg = errMsg[:500]
	}
	s.DB.Exec(`UPDATE source_status SET last_run_at=?, last_status=?, last_error=?, last_item_count=? WHERE source_key=?`,
		now(), status, errMsg, itemCount, key)
}

// HasSuccessRunToday 当日是否存在成功抓取（启动补抓判断）
func (s *Store) HasSuccessRunToday() bool {
	var n int
	today := time.Now().Format("2006-01-02")
	if s.DB.QueryRow(`SELECT COUNT(*) FROM source_status WHERE last_status='success' AND last_run_at >= ?`, today).Scan(&n) != nil {
		return false
	}
	return n > 0
}

// ---------- 维护 ----------

// Cleanup 清理过期业务数据
func (s *Store) Cleanup(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	for _, table := range []string{"news", "debt_item", "house_item"} {
		if _, err := s.DB.Exec(
			"DELETE FROM "+table+" WHERE fetched_at < ? AND id NOT IN (SELECT item_id FROM favorite WHERE item_type=?)",
			cutoff, tableToType(table)); err != nil {
			return err
		}
	}
	return nil
}

func tableToType(table string) string {
	switch table {
	case "news":
		return "news"
	case "debt_item":
		return "debt"
	case "house_item":
		return "house"
	}
	return ""
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
