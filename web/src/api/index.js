// API 封装：统一 token 注入 / 401 跳转 / code!=0 抛错

const TOKEN_KEY = 'wb_token'
const MUST_CHANGE_KEY = 'wb_must_change'

export const getToken = () => localStorage.getItem(TOKEN_KEY) || ''
export const setToken = (t) => localStorage.setItem(TOKEN_KEY, t)
export const getMustChange = () => localStorage.getItem(MUST_CHANGE_KEY) === '1'
export const setMustChange = (v) => localStorage.setItem(MUST_CHANGE_KEY, v ? '1' : '0')
export const clearToken = () => {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(MUST_CHANGE_KEY)
}

function qs(params) {
  const u = new URLSearchParams()
  Object.entries(params || {}).forEach(([k, v]) => {
    if (v === true) u.set(k, '1')
    else if (v === false || v === '' || v === undefined || v === null) return
    else u.set(k, v)
  })
  const s = u.toString()
  return s ? '?' + s : ''
}

async function request(method, url, body) {
  const headers = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(url, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  if (res.status === 401) {
    clearToken()
    if (location.pathname !== '/login') location.href = '/login'
    throw new Error('登录已过期，请重新登录')
  }
  let j
  try {
    j = await res.json()
  } catch {
    throw new Error(`服务异常（HTTP ${res.status}）`)
  }
  if (j.code !== 0) throw new Error(j.msg || '请求失败')
  return j.data
}

export const api = {
  // 认证
  login: (username, password) => request('POST', '/api/auth/login', { username, password }),
  changePassword: (old_password, new_password) =>
    request('POST', '/api/auth/change-password', { old_password, new_password }),

  // 新闻
  newsList: (params) => request('GET', '/api/news/list' + qs(params)),
  newsMarkRead: (ids) => request('POST', '/api/news/mark-read', { ids }),

  // 债权
  debtList: (params) => request('GET', '/api/debt/list' + qs(params)),
  debtMarkRead: (ids) => request('POST', '/api/debt/mark-read', { ids }),

  // 住宅法拍
  houseList: (params) => request('GET', '/api/house/list' + qs(params)),
  houseMarkRead: (ids) => request('POST', '/api/house/mark-read', { ids }),

  // 收藏
  favoriteList: (params) => request('GET', '/api/favorite/list' + qs(params)),
  favoriteAdd: (item_type, item_id, note) =>
    request('POST', '/api/favorite/add', { item_type, item_id, note }),
  favoriteRemove: (id) => request('POST', '/api/favorite/remove', { id }),

  // 统计
  statsOverview: () => request('GET', '/api/stats/overview'),

  // 数据源
  sourceList: () => request('GET', '/api/source/list'),
  sourceToggle: (source_key, enabled) => request('POST', '/api/source/toggle', { source_key, enabled }),
  sourceRun: (source_key) => request('POST', '/api/source/run', { source_key }),

  // 配置
  configList: () => request('GET', '/api/config/list'),
  configUpdate: (key, value) => request('POST', '/api/config/update', { key, value }),

  // 备份
  backupList: () => request('GET', '/api/backup/list'),
  backupRun: () => request('POST', '/api/backup/run', {}),
}
