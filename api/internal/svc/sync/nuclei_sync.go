package sync

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cscan/model"

	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/yaml.v3"
)

// NucleiSyncService Nuclei模板同步服务
type NucleiSyncService struct {
	model      *model.NucleiTemplateModel
	categories []string
	stats      map[string]int
}

// NewNucleiSyncService 创建同步服务
func NewNucleiSyncService(model *model.NucleiTemplateModel) *NucleiSyncService {
	return &NucleiSyncService{
		model:      model,
		categories: []string{},
		stats:      map[string]int{},
	}
}

// SyncTemplates 同步Nuclei模板到数据库
func (s *NucleiSyncService) SyncTemplates(ctx context.Context) {
	templatesDir := getNucleiTemplatesDir()
	if templatesDir == "" {
		logx.Info("[NucleiSync] Nuclei templates directory not found, skipping sync")
		return
	}

	logx.Infof("[NucleiSync] Starting sync from: %s", templatesDir)
	startTime := time.Now()

	var templates []*model.NucleiTemplate
	batchSize := 500
	totalCount := 0
	errorCount := 0

	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		template := parseNucleiTemplateFile(path, templatesDir)
		if template == nil {
			return nil
		}

		templates = append(templates, template)
		totalCount++

		if len(templates) >= batchSize {
			if err := s.model.BulkUpsert(ctx, templates); err != nil {
				errorCount++
			}
			templates = templates[:0]
		}

		return nil
	})

	if err != nil {
		logx.Errorf("[NucleiSync] Walk error: %v", err)
	}

	if len(templates) > 0 {
		if err := s.model.BulkUpsert(ctx, templates); err != nil {
			errorCount++
		}
	}

	duration := time.Since(startTime)
	logx.Infof("[NucleiSync] Completed: %d templates synced in %v (%d batch errors)", totalCount, duration, errorCount)

	s.RefreshCache(ctx)
}

// RefreshCache 刷新缓存
func (s *NucleiSyncService) RefreshCache(ctx context.Context) {
	categories, err := s.model.GetCategories(ctx)
	if err == nil {
		sort.Strings(categories)
		s.categories = categories
	}

	stats, err := s.model.GetStats(ctx)
	if err == nil {
		s.stats = stats
	}

	logx.Infof("[NucleiCache] Refreshed: %d categories", len(s.categories))
}

// GetCategories 获取分类
func (s *NucleiSyncService) GetCategories() []string {
	return s.categories
}

// GetStats 获取统计
func (s *NucleiSyncService) GetStats() map[string]int {
	return s.stats
}

// getNucleiTemplatesDir 获取Nuclei模板目录
func getNucleiTemplatesDir() string {
	homeDir, _ := os.UserHomeDir()
	possiblePaths := []string{
		filepath.Join(homeDir, "nuclei-templates"),
		filepath.Join(homeDir, ".local", "nuclei-templates"),
		filepath.Join(homeDir, ".nuclei-templates"),
		"/opt/nuclei-templates",
		"C:\\nuclei-templates",
	}

	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
	}
	return ""
}

// NucleiTemplateYAML 用于解析YAML的结构
type NucleiTemplateYAML struct {
	Id   string `yaml:"id"`
	Info struct {
		Name        string `yaml:"name"`
		Author      any    `yaml:"author"`
		Severity    string `yaml:"severity"`
		Description string `yaml:"description"`
		Tags        string `yaml:"tags"`
	} `yaml:"info"`
}

// parseNucleiTemplateFile 解析Nuclei模板文件
func parseNucleiTemplateFile(filePath, baseDir string) *model.NucleiTemplate {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	var info NucleiTemplateYAML
	if err := yaml.Unmarshal(data, &info); err != nil {
		return nil
	}

	if info.Id == "" {
		return nil
	}

	relPath, _ := filepath.Rel(baseDir, filePath)
	category := ""
	if parts := strings.Split(relPath, string(os.PathSeparator)); len(parts) > 1 {
		category = parts[0]
	}

	author := ""
	switch v := info.Info.Author.(type) {
	case string:
		author = v
	case []interface{}:
		var authors []string
		for _, a := range v {
			if s, ok := a.(string); ok {
				authors = append(authors, s)
			}
		}
		author = strings.Join(authors, ", ")
	}

	var tags []string
	if info.Info.Tags != "" {
		for _, tag := range strings.Split(info.Info.Tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	severity := strings.ToLower(info.Info.Severity)
	if severity == "" {
		severity = "unknown"
	}

	return &model.NucleiTemplate{
		TemplateId:  info.Id,
		Name:        info.Info.Name,
		Author:      author,
		Severity:    severity,
		Description: info.Info.Description,
		Tags:        tags,
		Category:    category,
		FilePath:    relPath,
		Content:     string(data),
		Enabled:     true,
	}
}
