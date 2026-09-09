package sources

// util.go 数据源共用工具：新闻分类 / 时间解析 / 金额解析

import (
	"strconv"
	"strings"
	"time"
)

// centralAMCs 5大AMC 关键词（中央）
var centralAMCs = []struct{ Keyword, Name string }{
	{"信达", "中国信达"},
	{"中信金融资产", "中信金融资产"},
	{"华融", "中信金融资产"},
	{"东方资产", "中国东方"},
	{"长城资产", "中国长城"},
	{"银河资产", "银河资产"},
}

// classifyNews 按标题关键词分类新闻
func classifyNews(title string) (category, org string) {
	for _, c := range centralAMCs {
		if strings.Contains(title, c.Keyword) {
			return "central_amc", c.Name
		}
	}
	if strings.Contains(title, "银行") {
		return "bank_transfer", ""
	}
	if strings.Contains(title, "资产") || strings.Contains(title, "AMC") ||
		strings.Contains(title, "不良") {
		return "local_amc", ""
	}
	return "industry", ""
}

var timeLayouts = []string{
	time.RFC1123,
	time.RFC1123Z,
	"Mon, 2 Jan 2006 15:04:05 MST",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05Z",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"2006/01/02",
	"2006年1月2日 15:04",  // 京东公告时间：2026年9月8日 19:41
	"2006-1-2 15:04:05", // 公拍网时间：2026-9-10 10:00:00
	"2006-1-2 15:04",
	"2006年1月2日15:04",
}

// normalizeDate 尝试多种格式解析时间，统一输出 YYYY-MM-DD HH:mm（失败返回原文）
func normalizeDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	for _, layout := range timeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02 15:04")
		}
	}
	return s
}

// parseAmount 从文本提取金额数字（如 "1,234.56万元" → 1234.56）
func parseAmount(s string) float64 {
	var b strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' {
			b.WriteRune(r)
		}
	}
	v, err := strconv.ParseFloat(b.String(), 64)
	if err != nil {
		return 0
	}
	return v
}
