package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cscan/api/internal/config"
	"cscan/model"
	"cscan/rpc/task/pb"
	"cscan/scheduler"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
)

// truncateError 截断错误信息，最多显示指定长度
func truncateError(err error, maxLen int) string {
	if err == nil {
		return ""
	}
	errStr := err.Error()
	if len(errStr) > maxLen {
		return errStr[:maxLen] + "..."
	}
	return errStr
}

type ServiceContext struct {
	Config              config.Config
	MongoClient         *mongo.Client
	MongoDB             *mongo.Database
	RedisClient         *redis.Client
	Scheduler           *scheduler.Scheduler
	TaskRpcClient       pb.TaskServiceClient
	UserModel           *model.UserModel
	WorkspaceModel      *model.WorkspaceModel
	ProfileModel        *model.TaskProfileModel
	TagMappingModel     *model.TagMappingModel
	CustomPocModel      *model.CustomPocModel
	NucleiTemplateModel *model.NucleiTemplateModel
	FingerprintModel    *model.FingerprintModel

	// 缓存的模板元数据
	TemplateCategories []string
	TemplateTags       []string
	TemplateStats      map[string]int
}

func NewServiceContext(c config.Config) *ServiceContext {
	// MongoDB连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(c.Mongo.Uri))
	if err != nil {
		panic(err)
	}

	mongoDB := mongoClient.Database(c.Mongo.DbName)

	// Redis连接
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Host,
		Password: c.Redis.Pass,
		DB:       0,
	})

	// 创建调度器
	sched := scheduler.NewScheduler(rdb)

	// 创建RPC客户端
	taskRpcClient := pb.NewTaskServiceClient(zrpc.MustNewClient(c.TaskRpc).Conn())

	svcCtx := &ServiceContext{
		Config:              c,
		MongoClient:         mongoClient,
		MongoDB:             mongoDB,
		RedisClient:         rdb,
		Scheduler:           sched,
		TaskRpcClient:       taskRpcClient,
		UserModel:           model.NewUserModel(mongoDB),
		WorkspaceModel:      model.NewWorkspaceModel(mongoDB),
		ProfileModel:        model.NewTaskProfileModel(mongoDB),
		TagMappingModel:     model.NewTagMappingModel(mongoDB),
		CustomPocModel:      model.NewCustomPocModel(mongoDB),
		NucleiTemplateModel: model.NewNucleiTemplateModel(mongoDB),
		FingerprintModel:    model.NewFingerprintModel(mongoDB),
		TemplateCategories:  []string{},
		TemplateTags:        []string{},
		TemplateStats:       map[string]int{},
	}

	// 先加载缓存的模板元数据
	svcCtx.RefreshTemplateCache()

	// 异步同步Nuclei模板到数据库
	go svcCtx.SyncNucleiTemplates()

	// 异步同步Wappalyzer指纹到数据库
	go svcCtx.SyncWappalyzerFingerprints()

	return svcCtx
}

// GetAssetModel 根据workspaceId获取资产模型
func (s *ServiceContext) GetAssetModel(workspaceId string) *model.AssetModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewAssetModel(s.MongoDB, workspaceId)
}

// GetMainTaskModel 根据workspaceId获取主任务模型
func (s *ServiceContext) GetMainTaskModel(workspaceId string) *model.MainTaskModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewMainTaskModel(s.MongoDB, workspaceId)
}

// GetVulModel 根据workspaceId获取漏洞模型
func (s *ServiceContext) GetVulModel(workspaceId string) *model.VulModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewVulModel(s.MongoDB, workspaceId)
}

// GetAssetHistoryModel 根据workspaceId获取资产历史模型
func (s *ServiceContext) GetAssetHistoryModel(workspaceId string) *model.AssetHistoryModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewAssetHistoryModel(s.MongoDB, workspaceId)
}


// SyncNucleiTemplates 同步Nuclei模板到数据库
func (s *ServiceContext) SyncNucleiTemplates() {
	ctx := context.Background()

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

		// 批量写入
		if len(templates) >= batchSize {
			if err := s.NucleiTemplateModel.BulkUpsert(ctx, templates); err != nil {
				errorCount++
			}
			templates = templates[:0]
		}

		return nil
	})

	if err != nil {
		logx.Errorf("[NucleiSync] Walk error: %s", truncateError(err, 200))
	}

	// 写入剩余的模板
	if len(templates) > 0 {
		if err := s.NucleiTemplateModel.BulkUpsert(ctx, templates); err != nil {
			errorCount++
		}
	}

	duration := time.Since(startTime)
	if errorCount > 0 {
		logx.Infof("[NucleiSync] Completed: %d templates synced in %v (%d batch errors ignored)", totalCount, duration, errorCount)
	} else {
		logx.Infof("[NucleiSync] Completed: %d templates synced in %v", totalCount, duration)
	}

	// 同步完成后刷新缓存
	s.RefreshTemplateCache()
}

// RefreshTemplateCache 刷新模板元数据缓存
func (s *ServiceContext) RefreshTemplateCache() {
	ctx := context.Background()

	// 获取分类
	categories, err := s.NucleiTemplateModel.GetCategories(ctx)
	if err == nil {
		sort.Strings(categories)
		s.TemplateCategories = categories
	}

	// 标签改为用户输入模糊查询，不再缓存
	s.TemplateTags = []string{}

	// 获取统计信息
	stats, err := s.NucleiTemplateModel.GetStats(ctx)
	if err == nil {
		s.TemplateStats = stats
	}

	logx.Infof("[NucleiCache] Refreshed: %d categories, stats: %v", len(s.TemplateCategories), s.TemplateStats)
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
	
	// 获取分类(第一级目录名)
	relPath, _ := filepath.Rel(baseDir, filePath)
	category := ""
	if parts := strings.Split(relPath, string(os.PathSeparator)); len(parts) > 1 {
		category = parts[0]
	}
	
	// 解析作者
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
	
	// 解析标签
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
		Content:     string(data), // 存储YAML内容
		Enabled:     true,         // 默认启用
	}
}


// SyncWappalyzerFingerprints 同步Wappalyzer指纹到数据库
func (s *ServiceContext) SyncWappalyzerFingerprints() {
	ctx := context.Background()

	// 检查是否已有内置指纹
	count, _ := s.FingerprintModel.Count(ctx, map[string]interface{}{"is_builtin": true})
	if count > 0 {
		logx.Infof("[FingerprintSync] Found %d builtin fingerprints, skipping sync", count)
		return
	}

	logx.Info("[FingerprintSync] Starting sync of Wappalyzer fingerprints...")
	startTime := time.Now()

	// 获取wappalyzergo内置的指纹数据
	fingerprintsData := wappalyzer.GetFingerprints()
	if fingerprintsData == "" {
		logx.Info("[FingerprintSync] No fingerprints data available")
		return
	}

	// 解析指纹数据 - Fingerprints结构包含Apps字段
	var fingerprints wappalyzer.Fingerprints
	if err := json.Unmarshal([]byte(fingerprintsData), &fingerprints); err != nil {
		logx.Errorf("[FingerprintSync] Failed to parse fingerprints: %s", truncateError(err, 200))
		return
	}

	// Wappalyzer分类映射（常见分类）
	categoryNames := map[int]string{
		1: "CMS", 2: "Message boards", 3: "Database managers", 4: "Documentation",
		5: "Widgets", 6: "Ecommerce", 7: "Photo galleries", 8: "Wikis",
		9: "Hosting panels", 10: "Analytics", 11: "Blogs", 12: "JavaScript frameworks",
		13: "Issue trackers", 14: "Video players", 15: "Comment systems", 16: "Security",
		17: "Font scripts", 18: "Web frameworks", 19: "Miscellaneous", 20: "Editors",
		21: "LMS", 22: "Web servers", 23: "Caching", 24: "Rich text editors",
		25: "JavaScript graphics", 26: "Mobile frameworks", 27: "Programming languages",
		28: "Operating systems", 29: "Search engines", 30: "Webmail", 31: "CDN",
		32: "Marketing automation", 33: "Web server extensions", 34: "Databases",
		35: "Maps", 36: "Advertising", 37: "Network devices", 38: "Media servers",
		39: "Webcams", 40: "Printers", 41: "Payment processors", 42: "Tag managers",
		43: "Paywalls", 44: "Build systems", 45: "Control systems", 46: "Remote access",
		47: "Dev tools", 48: "Network storage", 49: "Feed readers", 50: "DMS",
		51: "Page builders", 52: "Live chat", 53: "CRM", 54: "SEO", 55: "Accounting",
		56: "Cryptominers", 57: "Static site generator", 58: "User onboarding",
		59: "JavaScript libraries", 60: "Containers", 61: "SaaS", 62: "PaaS",
		63: "IaaS", 64: "Reverse proxies", 65: "Load balancers", 66: "UI frameworks",
		67: "Cookie compliance", 68: "Accessibility", 69: "Social login", 70: "SSL/TLS certificate authorities",
		71: "Affiliate programs", 72: "Appointment scheduling", 73: "Surveys", 74: "A/B Testing",
		75: "Email", 76: "Personalisation", 77: "Retargeting", 78: "RUM", 79: "Geolocation",
		80: "WordPress themes", 81: "Shopify themes", 82: "WordPress plugins", 83: "Shopify apps",
		84: "Drupal themes", 85: "Browser fingerprinting", 86: "Loyalty & rewards",
		87: "Feature management", 88: "Segmentation", 89: "Hosting", 90: "Translation",
		91: "Reviews", 92: "Buy now pay later", 93: "Performance", 94: "Reservations & delivery",
		95: "Referral marketing", 96: "Digital asset management", 97: "Content curation",
		98: "Customer data platform", 99: "Cart abandonment", 100: "Shipping carriers",
		101: "Fulfilment", 102: "Returns", 103: "Cross border ecommerce",
	}

	totalCount := 0
	for name, fp := range fingerprints.Apps {
		// 获取分类名称
		categoryName := "unknown"
		if len(fp.Cats) > 0 {
			if catName, ok := categoryNames[fp.Cats[0]]; ok {
				categoryName = catName
			} else {
				categoryName = fmt.Sprintf("category-%d", fp.Cats[0])
			}
		}
		
		// 转换Meta字段（map[string][]string -> map[string]string）
		metaMap := make(map[string]string)
		for k, v := range fp.Meta {
			if len(v) > 0 {
				metaMap[k] = strings.Join(v, " | ")
			}
		}

		// 转换Dom字段为JSON字符串
		domStr := ""
		if len(fp.Dom) > 0 {
			if domBytes, err := json.Marshal(fp.Dom); err == nil {
				domStr = string(domBytes)
			}
		}

		doc := &model.Fingerprint{
			Name:        name,
			Category:    categoryName,
			Website:     fp.Website,
			Icon:        fp.Icon,
			Description: fp.Description,
			IsBuiltin:   true,
			Enabled:     true,
			Headers:     fp.Headers,
			Cookies:     fp.Cookies,
			HTML:        fp.HTML,
			Scripts:     fp.Script,
			ScriptSrc:   fp.ScriptSrc,
			JS:          fp.JS,
			Meta:        metaMap,
			CSS:         fp.CSS,
			Dom:         domStr,
			Implies:     fp.Implies,
			CPE:         fp.CPE,
		}

		if err := s.FingerprintModel.Upsert(ctx, doc); err != nil {
			logx.Errorf("[FingerprintSync] Failed to upsert %s: %s", name, truncateError(err, 200))
			continue
		}
		totalCount++
	}

	duration := time.Since(startTime)
	logx.Infof("[FingerprintSync] Completed: %d fingerprints synced in %v", totalCount, duration)
}
