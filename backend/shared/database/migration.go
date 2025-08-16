package database

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

// Migration 数据库迁移管理器
type Migration struct {
	db gdb.DB
}

// NewMigration 创建迁移管理器实例
func NewMigration(db gdb.DB) *Migration {
	return &Migration{db: db}
}

// MigrationRecord 迁移记录
type MigrationRecord struct {
	ID        uint64    `json:"id" orm:"id"`
	Filename  string    `json:"filename" orm:"filename"`
	ExecutedAt time.Time `json:"executed_at" orm:"executed_at"`
}

// initMigrationTable 初始化迁移记录表
func (m *Migration) initMigrationTable(ctx context.Context) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS migrations (
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		filename varchar(255) NOT NULL UNIQUE,
		executed_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_filename (filename)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据库迁移记录表'`

	_, err := m.db.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("创建迁移记录表失败: %v", err)
	}
	return nil
}

// getExecutedMigrations 获取已执行的迁移记录
func (m *Migration) getExecutedMigrations(ctx context.Context) (map[string]bool, error) {
	executed := make(map[string]bool)
	
	records, err := m.db.Model("migrations").All()
	if err != nil {
		return nil, fmt.Errorf("查询迁移记录失败: %v", err)
	}
	
	for _, record := range records {
		executed[record["filename"].String()] = true
	}
	
	return executed, nil
}

// recordMigration 记录迁移执行
func (m *Migration) recordMigration(ctx context.Context, filename string) error {
	_, err := m.db.Model("migrations").Insert(map[string]interface{}{
		"filename":    filename,
		"executed_at": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("记录迁移失败: %v", err)
	}
	return nil
}

// RunMigrations 执行数据库迁移
func (m *Migration) RunMigrations(ctx context.Context, migrationDir string) error {
	// 初始化迁移记录表
	if err := m.initMigrationTable(ctx); err != nil {
		return err
	}
	
	// 获取已执行的迁移
	executed, err := m.getExecutedMigrations(ctx)
	if err != nil {
		return err
	}
	
	// 获取迁移文件列表
	files, err := gfile.ScanDirFile(migrationDir, "*.sql", false)
	if err != nil {
		return fmt.Errorf("扫描迁移目录失败: %v", err)
	}
	
	// 按文件名排序确保执行顺序
	sort.Strings(files)
	
	// 执行未运行的迁移
	for _, file := range files {
		filename := filepath.Base(file)
		
		// 跳过已执行的迁移
		if executed[filename] {
			g.Log().Infof(ctx, "跳过已执行的迁移: %s", filename)
			continue
		}
		
		// 读取并执行迁移文件
		if err := m.executeMigrationFile(ctx, file, filename); err != nil {
			return fmt.Errorf("执行迁移文件 %s 失败: %v", filename, err)
		}
		
		g.Log().Infof(ctx, "成功执行迁移: %s", filename)
	}
	
	return nil
}

// executeMigrationFile 执行单个迁移文件
func (m *Migration) executeMigrationFile(ctx context.Context, filepath, filename string) error {
	// 读取迁移文件内容
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("读取迁移文件失败: %v", err)
	}
	
	// 分割SQL语句（通过分号和换行符）
	sqls := strings.Split(string(content), ";")
	
	// 开始事务
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	
	// 执行每个SQL语句
	for _, sql := range sqls {
		sql = strings.TrimSpace(sql)
		if sql == "" || strings.HasPrefix(sql, "--") {
			continue
		}
		
		if _, err := tx.Exec(sql); err != nil {
			return fmt.Errorf("执行SQL失败 [%s]: %v", sql, err)
		}
	}
	
	// 记录迁移执行
	if err := m.recordMigration(ctx, filename); err != nil {
		return err
	}
	
	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}
	
	return nil
}

// GetMigrationStatus 获取迁移状态
func (m *Migration) GetMigrationStatus(ctx context.Context, migrationDir string) (map[string]bool, error) {
	// 获取已执行的迁移
	executed, err := m.getExecutedMigrations(ctx)
	if err != nil {
		return nil, err
	}
	
	// 获取所有迁移文件
	files, err := gfile.ScanDirFile(migrationDir, "*.sql", false)
	if err != nil {
		return nil, fmt.Errorf("扫描迁移目录失败: %v", err)
	}
	
	status := make(map[string]bool)
	for _, file := range files {
		filename := filepath.Base(file)
		status[filename] = executed[filename]
	}
	
	return status, nil
}