package store

// users.go 用户表访问

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type UserRow struct {
	ID           int64
	Username     string
	PasswordHash string
	MustChange   int
}

func (s *Store) CountUsers(ctx context.Context) (int64, error) {
	var n int64
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM user`).Scan(&n)
	return n, err
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string, mustChange int) error {
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO user(username, password_hash, must_change_password, created_at, updated_at)
		 VALUES(?,?,?,?,?)`, username, passwordHash, mustChange, now(), now())
	return err
}

func (s *Store) GetUser(ctx context.Context, username string) (*UserRow, error) {
	var u UserRow
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, username, password_hash, must_change_password FROM user WHERE username=?`,
		username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.MustChange)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) UpdatePassword(ctx context.Context, id int64, passwordHash string, mustChange int) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE user SET password_hash=?, must_change_password=?, updated_at=? WHERE id=?`,
		passwordHash, mustChange, now(), id)
	return err
}

// RandomPassword 生成 12 位随机十六进制初始密码
func RandomPassword() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
