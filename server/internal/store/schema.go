package store

// schema.go 建表迁移 + 配置种子（幂等，启动时执行）

const schemaDDL = `
CREATE TABLE IF NOT EXISTS user (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	must_change_password INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS news (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	source_key TEXT NOT NULL,
	url TEXT NOT NULL UNIQUE,
	title TEXT NOT NULL,
	summary TEXT NOT NULL DEFAULT '',
	category TEXT NOT NULL DEFAULT 'industry',
	org_name TEXT NOT NULL DEFAULT '',
	published_at TEXT NOT NULL DEFAULT '',
	fetched_at TEXT NOT NULL,
	is_read INTEGER NOT NULL DEFAULT 0,
	fingerprint TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_news_cat_pub ON news(category, published_at);
CREATE INDEX IF NOT EXISTS idx_news_read ON news(is_read);
CREATE INDEX IF NOT EXISTS idx_news_fp ON news(fingerprint);

CREATE TABLE IF NOT EXISTS debt_item (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	source_key TEXT NOT NULL,
	url TEXT NOT NULL UNIQUE,
	title TEXT NOT NULL,
	summary TEXT NOT NULL DEFAULT '',
	transferor TEXT NOT NULL DEFAULT '',
	amount_wan REAL NOT NULL DEFAULT 0,
	region TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT '',
	end_time TEXT NOT NULL DEFAULT '',
	published_at TEXT NOT NULL DEFAULT '',
	fetched_at TEXT NOT NULL,
	is_read INTEGER NOT NULL DEFAULT 0,
	fingerprint TEXT NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_debt_fp ON debt_item(fingerprint);
CREATE INDEX IF NOT EXISTS idx_debt_region ON debt_item(region);

CREATE TABLE IF NOT EXISTS house_item (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	source_key TEXT NOT NULL,
	url TEXT NOT NULL UNIQUE,
	title TEXT NOT NULL,
	city TEXT NOT NULL DEFAULT '',
	district TEXT NOT NULL DEFAULT '',
	area_sqm REAL NOT NULL DEFAULT 0,
	start_price_wan REAL NOT NULL DEFAULT 0,
	eval_price_wan REAL NOT NULL DEFAULT 0,
	court TEXT NOT NULL DEFAULT '',
	auction_date TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT '',
	published_at TEXT NOT NULL DEFAULT '',
	fetched_at TEXT NOT NULL,
	is_read INTEGER NOT NULL DEFAULT 0,
	fingerprint TEXT NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_house_fp ON house_item(fingerprint);
CREATE INDEX IF NOT EXISTS idx_house_city_auc ON house_item(city, auction_date);

CREATE TABLE IF NOT EXISTS favorite (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	item_type TEXT NOT NULL,
	item_id INTEGER NOT NULL,
	note TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	UNIQUE(item_type, item_id)
);

CREATE TABLE IF NOT EXISTS source_status (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	source_key TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	category TEXT NOT NULL,
	enabled INTEGER NOT NULL DEFAULT 1,
	last_run_at TEXT NOT NULL DEFAULT '',
	last_status TEXT NOT NULL DEFAULT '',
	last_error TEXT NOT NULL DEFAULT '',
	last_item_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS app_config (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL DEFAULT ''
);
`

// 默认配置种子
var configSeeds = map[string]string{
	"daily_reset.read_mark": "on", // is_read 每日 0 点重置
	"retention.days":        "90", // 业务条目保留天数
	"backup.keep":           "14", // 备份保留份数
}

func (s *Store) migrate() error {
	if _, err := s.DB.Exec(schemaDDL); err != nil {
		return err
	}
	for k, v := range configSeeds {
		if _, err := s.DB.Exec(
			`INSERT INTO app_config(key, value) VALUES(?, ?)
			 ON CONFLICT(key) DO NOTHING`, k, v); err != nil {
			return err
		}
	}
	return nil
}
