package backup

// backup.go SQLite 在线备份：VACUUM INTO + 滚动保留

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"workbench/server/internal/store"
)

// Run 执行一次备份（VACUUM INTO 为在线一致备份，无需停服）
func Run(st *store.Store) error {
	dir := filepath.Join(st.DataDir, "backup")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	target := filepath.Join(dir, "app-"+time.Now().Format("20060102")+".db")
	// VACUUM INTO 目标文件不能已存在
	if _, err := os.Stat(target); err == nil {
		_ = os.Remove(target)
	}
	if _, err := st.DB.Exec(fmt.Sprintf("VACUUM INTO '%s'", target)); err != nil {
		return fmt.Errorf("VACUUM INTO 失败: %w", err)
	}
	prune(dir, st.ConfigGetInt("backup.keep", 14))
	return nil
}

// Info 备份文件信息
type Info struct {
	Name string `json:"name"`
	Size int64  `json:"size"` // 字节
}

// List 列出备份文件（新→旧）
func List(st *store.Store) []Info {
	dir := filepath.Join(st.DataDir, "backup")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []Info{}
	}
	list := []Info{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "app-") || !strings.HasSuffix(e.Name(), ".db") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		list = append(list, Info{Name: e.Name(), Size: info.Size()})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name > list[j].Name })
	return list
}

// prune 按文件名（含日期）倒序保留前 keep 份
func prune(dir string, keep int) {
	if keep <= 0 {
		keep = 14
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	names := []string{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "app-") && strings.HasSuffix(e.Name(), ".db") {
			names = append(names, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	for _, name := range names[keep:] {
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			log.Printf("[backup] 删除旧备份失败 %s: %v", name, err)
		}
	}
}
