package api

// handlers.go 业务接口实现

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"workbench/server/internal/backup"
	"workbench/server/internal/scraper"
	"workbench/server/internal/store"
)

var errInvalidParam = errors.New("参数错误")

type handlers struct {
	st *store.Store
}

func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": data})
}

func fail(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
}

func queryInt(c *gin.Context, key string, def int64) int64 {
	v, err := strconv.ParseInt(c.Query(key), 10, 64)
	if err != nil || v < 1 {
		return def
	}
	return v
}

func pageOf(c *gin.Context) store.Page {
	return store.Page{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "page_size", 20)}
}

func boolOf(c *gin.Context, key string) bool {
	return c.Query(key) == "1" || c.Query(key) == "true"
}

// ---------- 新闻 / 债权 / 法拍 ----------

func (h *handlers) newsList(c *gin.Context) {
	total, list, err := h.st.ListNews(c.Request.Context(), store.NewsFilter{
		Category: c.Query("category"),
		Keyword:  c.Query("keyword"),
		Unread:   boolOf(c, "unread"),
		Page:     pageOf(c),
	})
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"total": total, "list": list})
}

func (h *handlers) newsMarkRead(c *gin.Context) { h.markRead(c, "news") }

func (h *handlers) debtList(c *gin.Context) {
	total, list, err := h.st.ListDebt(c.Request.Context(), store.DebtFilter{
		SourceKey: c.Query("source_key"),
		Region:    c.Query("region"),
		Status:    c.Query("status"),
		Keyword:   c.Query("keyword"),
		Page:      pageOf(c),
	})
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"total": total, "list": list})
}

func (h *handlers) debtMarkRead(c *gin.Context) { h.markRead(c, "debt") }

func (h *handlers) houseList(c *gin.Context) {
	total, list, err := h.st.ListHouse(c.Request.Context(), store.HouseFilter{
		City:     c.Query("city"),
		District: c.Query("district"),
		Status:   c.Query("status"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		Keyword:  c.Query("keyword"),
		Page:     pageOf(c),
	})
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"total": total, "list": list})
}

func (h *handlers) houseMarkRead(c *gin.Context) { h.markRead(c, "house") }

func (h *handlers) markRead(c *gin.Context, itemType string) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		fail(c, errInvalidParam)
		return
	}
	if err := h.st.MarkRead(c.Request.Context(), itemType, req.IDs); err != nil {
		fail(c, err)
		return
	}
	ok(c, "done")
}

// ---------- 收藏 ----------

func (h *handlers) favoriteList(c *gin.Context) {
	total, list, err := h.st.ListFavorites(c.Request.Context(), c.Query("item_type"), pageOf(c))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"total": total, "list": list})
}

func (h *handlers) favoriteAdd(c *gin.Context) {
	var req struct {
		ItemType string `json:"item_type"`
		ItemID   int64  `json:"item_id"`
		Note     string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ItemID < 1 {
		fail(c, errInvalidParam)
		return
	}
	if err := h.st.FavoriteAdd(c.Request.Context(), req.ItemType, req.ItemID, req.Note); err != nil {
		fail(c, err)
		return
	}
	ok(c, "done")
}

func (h *handlers) favoriteRemove(c *gin.Context) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID < 1 {
		fail(c, errInvalidParam)
		return
	}
	if err := h.st.FavoriteRemove(c.Request.Context(), req.ID); err != nil {
		fail(c, err)
		return
	}
	ok(c, "done")
}

// ---------- 统计 ----------

func (h *handlers) statsOverview(c *gin.Context) {
	o, err := h.st.StatsOverview(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, o)
}

// ---------- 数据源 ----------

func (h *handlers) sourceList(c *gin.Context) {
	list, err := h.st.ListSources(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, list)
}

func (h *handlers) sourceToggle(c *gin.Context) {
	var req struct {
		SourceKey string `json:"source_key"`
		Enabled   bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.SourceKey == "" {
		fail(c, errInvalidParam)
		return
	}
	if err := h.st.SourceToggle(c.Request.Context(), req.SourceKey, req.Enabled); err != nil {
		fail(c, err)
		return
	}
	ok(c, "done")
}

func (h *handlers) sourceRun(c *gin.Context) {
	var req struct {
		SourceKey string `json:"source_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.SourceKey == "" {
		fail(c, errInvalidParam)
		return
	}
	s, found := scraper.Get(req.SourceKey)
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "未知数据源"})
		return
	}
	go scraper.RunOne(h.st, s)
	ok(c, "已触发抓取，请稍后刷新查看结果")
}

// ---------- 配置 ----------

func (h *handlers) configList(c *gin.Context) {
	list, err := h.st.ConfigAll(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, list)
}

func (h *handlers) configUpdate(c *gin.Context) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Key == "" {
		fail(c, errInvalidParam)
		return
	}
	if err := h.st.ConfigSet(c.Request.Context(), req.Key, req.Value); err != nil {
		fail(c, err)
		return
	}
	ok(c, "done")
}

// ---------- 备份（数据管理页补充能力） ----------

func (h *handlers) backupList(c *gin.Context) {
	ok(c, backup.List(h.st))
}

func (h *handlers) backupRun(c *gin.Context) {
	go func() {
		if err := backup.Run(h.st); err != nil {
			_ = err // 失败记录在服务日志中
		}
	}()
	ok(c, "已触发备份")
}
