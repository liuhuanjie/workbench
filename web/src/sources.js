// 数据源平台展示名（对应后端 source_key）
const SOURCE_LABEL = {
  // 住宅法拍
  gpai_house: '公拍网',
  jd_house: '京东司法拍卖',
  ali_house: '阿里资产',
  // 债权挂牌
  cex_sh: '上海联交所',
  cex_bj: '北京产权交易所',
  cex_gd: '广东联合产权',
  // 新闻
  cinda_news: '中国信达',
  citic_amc_news: '中信金融资产',
  orient_news: '中国东方',
  greatwall_news: '中国长城',
  galaxy_news: '银河资产',
  industry_rss: '行业媒体',
}

export function sourceLabel(key) {
  return SOURCE_LABEL[key] || key || '—'
}
