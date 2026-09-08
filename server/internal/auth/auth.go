package auth

// auth.go JWT 认证：登录/改密/中间件/防爆破/admin 初始化

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"workbench/server/internal/store"
)

const (
	failWindow  = 5 * time.Minute  // 计数窗口
	failLimit   = 5                // 窗口内允许失败次数
	lockTimeout = 10 * time.Minute // 锁定时长
)

type Service struct {
	st     *store.Store
	secret []byte
	mu     sync.Mutex
	fails  map[string]*failRecord
}

type failRecord struct {
	count       int
	windowStart time.Time
	lockedUntil time.Time
}

func NewService(st *store.Store, secret string) *Service {
	return &Service{st: st, secret: []byte(secret), fails: map[string]*failRecord{}}
}

// EnsureAdmin 首次启动初始化 admin 账号（随机密码打印到日志）
func (sv *Service) EnsureAdmin() {
	ctx := context.Background()
	n, err := sv.st.CountUsers(ctx)
	if err != nil {
		log.Fatalf("查询用户表失败: %v", err)
	}
	if n > 0 {
		return
	}
	pw := store.RandomPassword()
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), 10)
	if err != nil {
		log.Fatalf("生成密码哈希失败: %v", err)
	}
	if err := sv.st.CreateUser(ctx, "admin", string(hash), 1); err != nil {
		log.Fatalf("初始化 admin 失败: %v", err)
	}
	log.Println("==================================================")
	log.Println("  初始账号: admin")
	log.Println("  初始密码:", pw)
	log.Println("  （首次登录强制修改密码，请立即记录）")
	log.Println("==================================================")
}

// Login 登录接口
func (sv *Service) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	if err := sv.checkLimit(c.ClientIP(), req.Username); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	u, err := sv.st.GetUser(c.Request.Context(), req.Username)
	if err == nil && bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) == nil {
		sv.clearFail(c.ClientIP(), req.Username)
		token, err := sv.issueToken(u.ID, u.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "签发令牌失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok",
			"data": gin.H{"token": token, "must_change_password": u.MustChange == 1}})
		return
	}
	sv.recordFail(c.ClientIP(), req.Username)
	c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "用户名或密码错误"})
}

// ChangePassword 修改密码
func (sv *Service) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.OldPassword == "" || len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误（新密码至少 8 位）"})
		return
	}
	username := c.GetString("username")
	u, err := sv.st.GetUser(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.OldPassword)) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "原密码错误"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "生成密码哈希失败"})
		return
	}
	if err := sv.st.UpdatePassword(c.Request.Context(), u.ID, string(hash), 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新密码失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": "密码已更新"})
}

// Middleware JWT 校验中间件
func (sv *Service) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("非法签名算法")
			}
			return sv.secret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "令牌无效或已过期"})
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if username, ok := claims["username"].(string); ok {
				c.Set("username", username)
			}
			if uid, ok := claims["uid"].(float64); ok {
				c.Set("uid", int64(uid))
			}
		}
		c.Next()
	}
}

func (sv *Service) issueToken(uid int64, username string) (string, error) {
	claims := jwt.MapClaims{
		"uid":      uid,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(sv.secret)
}

// ---------- 登录防爆破（内存计数） ----------

func limitKey(ip, username string) string { return ip + "|" + username }

func (sv *Service) checkLimit(ip, username string) error {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	r, ok := sv.fails[limitKey(ip, username)]
	if !ok {
		return nil
	}
	if time.Now().Before(r.lockedUntil) {
		return fmt.Errorf("尝试过于频繁，请 10 分钟后再试")
	}
	return nil
}

func (sv *Service) recordFail(ip, username string) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	key := limitKey(ip, username)
	r, ok := sv.fails[key]
	if !ok || time.Since(r.windowStart) > failWindow {
		r = &failRecord{windowStart: time.Now()}
		sv.fails[key] = r
	}
	r.count++
	if r.count >= failLimit {
		r.lockedUntil = time.Now().Add(lockTimeout)
		r.count = 0
	}
}

func (sv *Service) clearFail(ip, username string) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	delete(sv.fails, limitKey(ip, username))
}
