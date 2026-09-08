# 不良资产投资工作台

个人不良资产投资研究工作台：聚合 AMC 动态、债权处置、住宅法拍信息，提供收藏/统计/数据管理能力。为 **2核2G 小型云服务器** 优化，支持镜像级备份与整站迁移。

## 功能概览

- 🏠 **工作台首页**：时钟、每日中英激励语、今日新增速览
- 📰 **业务新闻**：5大AMC / 地方AMC / 银行资产包转让 / 行业要闻，关键词分类与搜索
- 📋 **每日债权检索**：产交所 / 阿里 / 京东债权处置信息
- 🏘️ **每日住宅检索**：上海 / 北京住宅法拍（区域/面积/起拍价/开拍时间筛选）
- ⭐ **我的收藏**：跨类型收藏 + 备注，永不清理
- 📊 **数据统计**：未读数、收藏数、近 7 日新增趋势图
- 🗄️ **数据管理**：数据源健康状态、立即抓取、每日重置开关、备份管理
- ⚙️ **个人设置**：改密（首次登录强制）、退出

## 技术栈与版本要求

### 运行环境（服务器）

| 组件 | 版本要求 | 说明 |
|------|---------|------|
| Docker Engine | ≥ 20.10 | 部署机唯一依赖 |
| Docker Compose | v2（`docker compose` 子命令） | 编排两个容器 |
| 系统资源 | 2核2G 即可 | 整机占用约 410MB 内存；建议配 2G swap |
| 开放端口 | 80（HTTP） | HTTPS 已预留配置位，有域名后可启用 |

> 服务器上**不需要**安装 Go / Node / SQLite——全部封装在 Docker 镜像内。

### 构建 base 镜像版本（已固定在 Dockerfile 中）

| 镜像 | 用途 |
|------|------|
| `golang:1.22-alpine` | 后端构建（**Go 1.22**，CGO_ENABLED=0 纯静态编译） |
| `node:20-alpine` | 前端构建（Node 20 + npm） |
| `nginx:1.27-alpine` | 前端静态托管 + API 反向代理 |
| `alpine:3.20` | 后端运行镜像（含 tzdata，时区 Asia/Shanghai） |

### Go 依赖（server/go.mod）

| 依赖 | 用途 |
|------|------|
| `github.com/gin-gonic/gin` | Web 框架 |
| `github.com/golang-jwt/jwt/v5` | JWT 认证（HS256，24h 有效期） |
| `golang.org/x/crypto` | bcrypt 密码哈希 |
| `github.com/robfig/cron/v3` | 定时任务调度 |
| `modernc.org/sqlite` | 纯 Go SQLite 驱动（免 CGO，单文件数据库） |
| `github.com/PuerkitoBio/goquery` | HTML 列表页解析（数据源抓取） |

### 前端依赖（web/package.json）

`vue@3` · `vue-router@4` · `pinia` · `naive-ui` · `@iconify/vue`（fluent-emoji 可爱图标集）· `vite@5`（构建工具）

## 项目结构

```
workbench/
├── docker-compose.yml        # 编排：web(nginx) + app(go) + data 卷
├── Dockerfile.web            # 前端多阶段构建
├── Dockerfile.server         # 后端多阶段构建
├── .env.example              # 环境变量模板（复制为 .env）
├── nginx/nginx.conf          # 反代配置（HTTPS 预留位）
├── server/                   # Go 后端
│   └── internal/
│       ├── store/            # SQLite：建表、三层去重写入、全部查询
│       ├── auth/             # JWT + bcrypt + 登录防爆破 + admin 初始化
│       ├── api/              # REST 接口（路径全静态，参数走 query/body）
│       ├── scraper/          # 抓取器插件框架
│       │   └── sources/      # 各数据源（部分待上线校准，见文件内 TODO）
│       ├── cronjob/          # 每日 8/12/18 点抓取、0:05 已读重置、3:00 清理、4:00 备份
│       └── backup/           # VACUUM INTO 在线备份，滚动保留 14 份
├── web/                      # Vue3 前端（8 菜单布局，蓝绿清新主题）
└── data/                     # 运行时数据（挂载卷，含 app.db 与 backup/）
```

## 完整部署流程（阿里云 ECS 2c2g）

### 1. 安装 Docker

```bash
# 使用阿里云镜像源安装（以 CentOS/Alibaba Cloud Linux 为例）
dnf install -y dnf-plugin-releasever-adapter || true
dnf install -y docker docker-compose-plugin
systemctl enable --now docker

# 配置镜像加速（可选，加速拉取基础镜像）
mkdir -p /etc/docker
cat > /etc/docker/daemon.json <<'EOF'
{ "registry-mirrors": ["https://docker.mirrors.sjtug.sjtu.edu.cn"] }
EOF
systemctl restart docker
```

> Ubuntu 系统把 `dnf` 换成 `apt`，安装 `docker.io docker-compose-v2`。

### 2. 上传项目

```bash
# 方式一：git（推荐，便于后续升级）
git clone <你的仓库地址> workbench && cd workbench

# 方式二：本地打包上传（排除 node_modules）
# 本地: tar czf workbench.tar.gz --exclude node_modules --exclude data workbench/
# 然后: scp workbench.tar.gz root@<服务器IP>:~/ && 服务器上解压
```

### 3. 配置环境变量

```bash
cp .env.example .env
# 生成随机 JWT 密钥并写入
sed -i "s/change-me-to-a-random-secret/$(openssl rand -hex 32)/" .env
cat .env   # 确认 JWT_SECRET 已替换
```

### 4. 构建并启动

```bash
docker compose up -d --build
```

首次构建需拉取基础镜像并编译前后端，约 3~10 分钟。查看构建过程是否有报错：

```bash
docker compose build 2>&1 | tail -20
```

### 5. 获取初始密码

```bash
docker compose logs app | grep -A3 初始密码
```

输出形如：

```
==================================================
  初始账号: admin
  初始密码: a1b2c3d4e5f6
  （首次登录强制修改密码，请立即记录）
==================================================
```

### 6. 放行端口并访问

- 阿里云控制台 → ECS → 安全组 → 添加入方向规则：**TCP 80，源 0.0.0.0/0**（或限定你的常用 IP 更安全）
- 浏览器访问 `http://<公网IP>`，用 admin + 初始密码登录，按提示修改密码后即可使用

### 7. 建议的系统调优（可选）

```bash
# 开 2G swap，防抓取高峰 OOM
fallocate -l 2G /swapfile && chmod 600 /swapfile
mkswap /swapfile && swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab
```

## 日常运维

| 操作 | 命令 |
|------|------|
| 查看服务状态 | `docker compose ps` |
| 查看后端日志（含抓取结果） | `docker compose logs -f app` |
| 重启服务 | `docker compose restart` |
| 升级版本（代码更新后） | `git pull && docker compose up -d --build`（数据在 data/ 卷不受影响） |
| 手动触发某数据源抓取 | 数据管理页点“立即抓取”，或 `docker compose logs app` 观察输出 |

## 数据备份与迁移

### 自动备份

每日 04:00 自动执行 SQLite 在线备份至 `data/backup/app-YYYYMMDD.db`，滚动保留 14 份（数据管理页可查看列表、手动触发）。

### 整站迁移（退租换机时）

```bash
# ===== 旧服务器：导出 =====
docker save workbench-web:latest workbench-app:latest | gzip > workbench-images.tar.gz
tar czf workbench-data.tar.gz data/ .env

# ===== 传输（scp / 网盘 / OSS 均可）=====
scp workbench-*.tar.gz user@新服务器:~/

# ===== 新服务器：恢复 =====
mkdir -p ~/workbench && cd ~/workbench
docker load < workbench-images.tar.gz
tar xzf ~/workbench-data.tar.gz
docker compose up -d        # 服务 + 数据完整恢复
```

> 镜像可随时从源码重建，**真正不可再生的只有 `data/` 目录**——极端情况下手握一个 `app.db` 即可恢复全部数据。建议每半年做一次迁移演练。

## 数据源校准（上线后必做）

框架已就绪，但以下数据源的 URL / 选择器 / 接口**必须实测后填写**才能出数据（当前显示"失败-待校准"）：

| 文件 | 数据源 | 校准内容 |
|------|--------|---------|
| `server/internal/scraper/sources/news_rss.go` | 行业媒体 RSS | 填入实际 RSS/Atom 订阅地址（填好即生效） |
| `server/internal/scraper/sources/amc_news.go` | 5大AMC 官网 | 各官网新闻页 listURL 与 CSS 选择器 |
| `server/internal/scraper/sources/cex_debt.go` | 产交所挂牌 | 各所债权栏目 URL 与选择器 |
| `server/internal/scraper/sources/ali_house.go` | 阿里资产法拍 | 实测列表接口与反爬参数后实现解析 |
| `server/internal/scraper/sources/jd_house.go` | 京东法拍 | 实测接口后实现解析 |

校准方法：浏览器 F12 观察目标页面请求 → 填写配置 → `docker compose up -d --build` 重新构建 → 数据管理页点"立即抓取"验证。

## 常见问题

**Q: 构建时报 `npm ci` 失败？**
确认 `web/package-lock.json` 存在且与 `package.json` 同步（本地执行过 `npm install` 后提交）。

**Q: 80 端口被占用？**
修改 `docker-compose.yml` 中 web 服务的端口映射，如 `"8080:80"`，安全组同步放行新端口。

**Q: 忘记 admin 密码？**
在服务器上删除用户记录，重启后会重新生成初始密码（业务数据不受影响）：

```bash
docker compose exec app sqlite3 /app/data/app.db "DELETE FROM user;"
docker compose restart app
docker compose logs app | grep 初始密码
```

**Q: 如何启用 HTTPS？**
准备域名与证书后，放开 `nginx/nginx.conf` 中注释的 443 server 块，挂载证书目录并在 compose 中映射 443 端口，重新 `docker compose up -d --build`。

## 内存占用参考（2G 机器）

| 进程 | 内存 |
|------|------|
| 系统 + dockerd | ~350MB |
| Nginx 容器 | ~10MB |
| Go 应用容器 | ~50MB |
| **合计** | **~410MB / 2G**，余量充足 |
