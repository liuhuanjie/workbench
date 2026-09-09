// 图标统一出口：使用 @vicons/ionicons5（构建时打包，不走运行时 CDN）
// 说明：原先的 @iconify/vue 会在运行时向 api.iconify.design 请求图标数据，
// 国内访问慢导致页面每次加载多出 1~2 秒。改为本地打包后无网络请求。
export {
  CubeOutline,
  HomeOutline,
  NewspaperOutline,
  ReceiptOutline,
  BusinessOutline,
  BookmarkOutline,
  StatsChartOutline,
  ServerOutline,
  SettingsOutline,
  RefreshOutline,
  CloseOutline,
  StarOutline,
  Star,
  BulbOutline,
  DownloadOutline,
  PlayOutline,
} from '@vicons/ionicons5'
