package store

import (
	"database/sql"
	"os"
	"path/filepath"
)

// Store SQLite 数据访问层，所有 SQL 集中于此
type Store struct {
	DB      *sql.DB
	DataDir string
}

// Open 打开数据库并执行建表迁移
func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	dsn := "file:" + filepath.Join(dataDir, "app.db") +
		"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite 单文件库，连接池不宜过大
	db.SetMaxOpenConns(4)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &Store{DB: db, DataDir: dataDir}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.DB.Close() }
