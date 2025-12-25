package types

// ==================== 通用类型 ====================
type BaseResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type PageReq struct {
	Page     int `json:"page,default=1"`
	PageSize int `json:"pageSize,default=20"`
}

// ==================== 用户认证 ====================
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResp struct {
	Code        int    `json:"code"`
	Msg         string `json:"msg"`
	Token       string `json:"token"`
	UserId      string `json:"userId"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	WorkspaceId string `json:"workspaceId"`
}

type UserInfo struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
}

type UserListResp struct {
	Code  int        `json:"code"`
	Msg   string     `json:"msg"`
	Total int        `json:"total"`
	List  []UserInfo `json:"list"`
}

// ==================== 用户管理 ====================
type UserCreateReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Status   string `json:"status"`
}

type UserUpdateReq struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
}

type UserDeleteReq struct {
	Id string `json:"id"`
}

type UserResetPasswordReq struct {
	Id          string `json:"id"`
	NewPassword string `json:"newPassword"`
}

// ==================== 工作空间 ====================
type Workspace struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreateTime  string `json:"createTime"`
}

type WorkspaceListResp struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Total int         `json:"total"`
	List  []Workspace `json:"list"`
}

type WorkspaceSaveReq struct {
	Id          string `json:"id,optional"`
	Name        string `json:"name"`
	Description string `json:"description,optional"`
}

// ==================== 资产管理 ====================
type Asset struct {
	Id         string   `json:"id"`
	Authority  string   `json:"authority"`
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	Category   string   `json:"category"`
	Service    string   `json:"service"`
	Title      string   `json:"title"`
	App        []string `json:"app"`
	HttpStatus string   `json:"httpStatus"`
	HttpHeader string   `json:"httpHeader"`
	HttpBody   string   `json:"httpBody"`
	Banner     string   `json:"banner"`
	IconHash   string   `json:"iconHash"`
	Screenshot string   `json:"screenshot"`
	Location   string   `json:"location"`
	IsCDN      bool     `json:"isCdn"`
	IsCloud    bool     `json:"isCloud"`
	IsNew      bool     `json:"isNew"`
	IsUpdated  bool     `json:"isUpdated"`
	CreateTime string   `json:"createTime"`
	UpdateTime string   `json:"updateTime"`
	// 新增字段 - 风险评分
	RiskScore float64 `json:"riskScore,omitempty"`
	RiskLevel string  `json:"riskLevel,omitempty"`
}

type AssetListReq struct {
	Page         int    `json:"page,default=1"`
	PageSize     int    `json:"pageSize,default=20"`
	Query        string `json:"query,optional"`
	Host         string `json:"host,optional"`
	Port         int    `json:"port,optional"`
	Service      string `json:"service,optional"`
	Title        string `json:"title,optional"`
	App          string `json:"app,optional"`
	HttpStatus   string `json:"httpStatus,optional"`
	OnlyNew      bool   `json:"onlyNew,optional"`
	OnlyUpdated  bool   `json:"onlyUpdated,optional"`
	ExcludeCdn   bool   `json:"excludeCdn,optional"`
	SortByUpdate bool   `json:"sortByUpdate,optional"`
	// 新增字段 - 按风险评分排序
	SortByRisk bool `json:"sortByRisk,optional"`
}

type AssetListResp struct {
	Code  int     `json:"code"`
	Msg   string  `json:"msg"`
	Total int     `json:"total"`
	List  []Asset `json:"list"`
}

type AssetStatResp struct {
	Code         int        `json:"code"`
	Msg          string     `json:"msg"`
	TotalAsset   int        `json:"totalAsset"`
	TotalHost    int        `json:"totalHost"`
	NewCount     int        `json:"newCount"`
	UpdatedCount int        `json:"updatedCount"`
	TopPorts     []StatItem `json:"topPorts"`
	TopService   []StatItem `json:"topService"`
	TopApp       []StatItem `json:"topApp"`
	TopTitle     []StatItem `json:"topTitle"`
	// 新增字段 - 风险等级分布
	RiskDistribution map[string]int `json:"riskDistribution,omitempty"`
}

type StatItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type AssetDeleteReq struct {
	Id string `json:"id"`
}

type AssetBatchDeleteReq struct {
	Ids []string `json:"ids"`
}

type AssetHistoryReq struct {
	AssetId string `json:"assetId"`
	Limit   int    `json:"limit,default=20"`
}

type AssetHistoryItem struct {
	Id         string   `json:"id"`
	Authority  string   `json:"authority"`
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	Service    string   `json:"service"`
	Title      string   `json:"title"`
	App        []string `json:"app"`
	HttpStatus string   `json:"httpStatus"`
	HttpHeader string   `json:"httpHeader"`
	HttpBody   string   `json:"httpBody"`
	Banner     string   `json:"banner"`
	IconHash   string   `json:"iconHash"`
	Screenshot string   `json:"screenshot"`
	TaskId     string   `json:"taskId"`
	CreateTime string   `json:"createTime"`
}

type AssetHistoryResp struct {
	Code int                `json:"code"`
	Msg  string             `json:"msg"`
	List []AssetHistoryItem `json:"list"`
}

// ==================== 任务管理 ====================
type MainTask struct {
	Id           string `json:"id"`
	TaskId       string `json:"taskId"` // UUID，用于日志查询
	Name         string `json:"name"`
	Target       string `json:"target"`
	ProfileId    string `json:"profileId"`
	ProfileName  string `json:"profileName"`
	Status       string `json:"status"`
	Progress     int    `json:"progress"`
	Result       string `json:"result"`
	IsCron       bool   `json:"isCron"`
	CronRule     string `json:"cronRule"`
	CreateTime   string `json:"createTime"`
	SubTaskCount int    `json:"subTaskCount"` // 子任务总数
	SubTaskDone  int    `json:"subTaskDone"`  // 已完成子任务数
}

type MainTaskListReq struct {
	Page     int    `json:"page,default=1"`
	PageSize int    `json:"pageSize,default=20"`
	Name     string `json:"name,optional"`
	Status   string `json:"status,optional"`
}

type MainTaskListResp struct {
	Code  int        `json:"code"`
	Msg   string     `json:"msg"`
	Total int        `json:"total"`
	List  []MainTask `json:"list"`
}

type MainTaskCreateReq struct {
	Name      string `json:"name"`
	Target    string `json:"target"`
	ProfileId string `json:"profileId"`
	IsCron    bool   `json:"isCron,optional"`
	CronRule  string `json:"cronRule,optional"`
}

type TaskProfile struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      string `json:"config"`
}

type TaskProfileListResp struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	List []TaskProfile `json:"list"`
}

type TaskProfileSaveReq struct {
	Id          string `json:"id,optional"`
	Name        string `json:"name"`
	Description string `json:"description,optional"`
	Config      string `json:"config"`
}

type TaskProfileDeleteReq struct {
	Id string `json:"id"`
}

type MainTaskDeleteReq struct {
	Id string `json:"id"`
}

type MainTaskBatchDeleteReq struct {
	Ids []string `json:"ids"`
}

type MainTaskRetryReq struct {
	Id string `json:"id"`
}

type MainTaskControlReq struct {
	Id string `json:"id"`
}

// MainTaskUpdateReq 更新任务请求
type MainTaskUpdateReq struct {
	Id        string `json:"id"`                  // 任务ID
	Name      string `json:"name,optional"`       // 任务名称
	Target    string `json:"target,optional"`     // 扫描目标
	ProfileId string `json:"profileId,optional"`  // 配置ID
}

// GetTaskLogsReq 获取任务日志请求
type GetTaskLogsReq struct {
	TaskId string `json:"taskId"` // 任务ID
	Limit  int    `json:"limit,default=100"` // 返回条数限制
}

// TaskLogEntry 任务日志条目
type TaskLogEntry struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	WorkerName string `json:"workerName"`
	TaskId     string `json:"taskId"`
	Message    string `json:"message"`
}

// GetTaskLogsResp 获取任务日志响应
type GetTaskLogsResp struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	List []TaskLogEntry `json:"list"`
}

// ==================== 漏洞管理 ====================
type Vul struct {
	Id         string `json:"id"`
	Authority  string `json:"authority"`
	Url        string `json:"url"`
	PocFile    string `json:"pocFile"`
	Source     string `json:"source"`
	Severity   string `json:"severity"`
	Result     string `json:"result"`
	CreateTime string `json:"createTime"`
	// 新增字段 - 时间追踪
	FirstSeenTime string `json:"firstSeenTime,omitempty"`
	LastSeenTime  string `json:"lastSeenTime,omitempty"`
	ScanCount     int    `json:"scanCount,omitempty"`
}

// VulEvidence 漏洞证据链
type VulEvidence struct {
	MatcherName       string   `json:"matcherName,omitempty"`
	ExtractedResults  []string `json:"extractedResults,omitempty"`
	CurlCommand       string   `json:"curlCommand,omitempty"`
	Request           string   `json:"request,omitempty"`
	Response          string   `json:"response,omitempty"`
	ResponseTruncated bool     `json:"responseTruncated,omitempty"`
}

// VulDetail 漏洞详情（包含知识库信息和证据链）
type VulDetail struct {
	Id         string `json:"id"`
	Authority  string `json:"authority"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Url        string `json:"url"`
	PocFile    string `json:"pocFile"`
	Source     string `json:"source"`
	Severity   string `json:"severity"`
	Result     string `json:"result"`
	CreateTime string `json:"createTime"`
	// 知识库信息
	CvssScore   float64  `json:"cvssScore,omitempty"`
	CveId       string   `json:"cveId,omitempty"`
	CweId       string   `json:"cweId,omitempty"`
	Remediation string   `json:"remediation,omitempty"`
	References  []string `json:"references,omitempty"`
	// 证据链
	Evidence *VulEvidence `json:"evidence,omitempty"`
	// 时间追踪 
	FirstSeenTime string `json:"firstSeenTime,omitempty"`
	LastSeenTime  string `json:"lastSeenTime,omitempty"`
	ScanCount     int    `json:"scanCount,omitempty"`
}

// VulDetailReq 漏洞详情请求
type VulDetailReq struct {
	Id string `json:"id"`
}

// VulDetailResp 漏洞详情响应
type VulDetailResp struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data *VulDetail `json:"data,omitempty"`
}

type VulListReq struct {
	Page      int    `json:"page,default=1"`
	PageSize  int    `json:"pageSize,default=20"`
	Authority string `json:"authority,optional"`
	Severity  string `json:"severity,optional"`
	Source    string `json:"source,optional"`
}

type VulListResp struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Total int    `json:"total"`
	List  []Vul  `json:"list"`
}

type VulDeleteReq struct {
	Id string `json:"id"`
}

type VulBatchDeleteReq struct {
	Ids []string `json:"ids"`
}

// VulStatResp 漏洞统计响应
type VulStatResp struct {
	Code     int `json:"code"`
	Msg      string `json:"msg"`
	Total    int `json:"total"`
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
	Week     int `json:"week"`   // 近7天
	Month    int `json:"month"`  // 近30天
}

// TaskStatResp 任务统计响应
type TaskStatResp struct {
	Code      int `json:"code"`
	Msg       string `json:"msg"`
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Running   int `json:"running"`
	Failed    int `json:"failed"`
	Pending   int `json:"pending"`
	// 近7天每日趋势
	TrendDays      []string `json:"trendDays"`      // 日期标签
	TrendCompleted []int    `json:"trendCompleted"` // 每日完成数
	TrendFailed    []int    `json:"trendFailed"`    // 每日失败数
}

// ==================== Worker管理 ====================
type Worker struct {
	Name         string  `json:"name"`
	IP           string  `json:"ip"`
	CPULoad      float64 `json:"cpuLoad"`
	MemUsed      float64 `json:"memUsed"`
	TaskCount    int     `json:"taskCount"`    // 已执行任务数
	RunningCount int     `json:"runningCount"` // 正在执行任务数
	Status       string  `json:"status"`
	UpdateTime   string  `json:"updateTime"`
}

type WorkerListResp struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	List []Worker `json:"list"`
}

type WorkerDeleteReq struct {
	Name string `json:"name"` // Worker名称
}

type WorkerDeleteResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type WorkerRenameReq struct {
	OldName string `json:"oldName"` // 原Worker名称
	NewName string `json:"newName"` // 新Worker名称
}

type WorkerRenameResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// ==================== 在线API搜索 ====================
type OnlineSearchReq struct {
	Platform string `json:"platform"` // fofa/hunter/quake
	Query    string `json:"query"`
	Page     int    `json:"page,default=1"`
	PageSize int    `json:"pageSize,default=20"`
}

type OnlineSearchResult struct {
	Host     string `json:"host"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Domain   string `json:"domain"`
	Title    string `json:"title"`
	Server   string `json:"server"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Banner   string `json:"banner"`
	ICP      string `json:"icp"`
	Product  string `json:"product"`
	OS       string `json:"os"`
}

type OnlineSearchResp struct {
	Code  int                  `json:"code"`
	Msg   string               `json:"msg"`
	Total int                  `json:"total"`
	List  []OnlineSearchResult `json:"list"`
}

type OnlineImportReq struct {
	Assets []OnlineSearchResult `json:"assets"`
}

// ==================== API配置 ====================
type APIConfig struct {
	Id         string `json:"id"`
	Platform   string `json:"platform"`
	Key        string `json:"key"`
	Secret     string `json:"secret"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
}

type APIConfigListResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	List []APIConfig `json:"list"`
}

type APIConfigSaveReq struct {
	Id       string `json:"id,optional"`
	Platform string `json:"platform"`
	Key      string `json:"key"`
	Secret   string `json:"secret,optional"`
}


// ==================== POC标签映射 ====================
type TagMapping struct {
	Id          string   `json:"id"`
	AppName     string   `json:"appName"`
	NucleiTags  []string `json:"nucleiTags"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	CreateTime  string   `json:"createTime"`
}

type TagMappingListResp struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	List []TagMapping `json:"list"`
}

type TagMappingSaveReq struct {
	Id          string   `json:"id,optional"`
	AppName     string   `json:"appName"`
	NucleiTags  []string `json:"nucleiTags"`
	Description string   `json:"description,optional"`
	Enabled     bool     `json:"enabled"`
}

type TagMappingDeleteReq struct {
	Id string `json:"id"`
}

// ==================== 自定义POC ====================
type CustomPoc struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	TemplateId  string   `json:"templateId"`
	Severity    string   `json:"severity"`
	Tags        []string `json:"tags"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Enabled     bool     `json:"enabled"`
	CreateTime  string   `json:"createTime"`
}

type CustomPocListReq struct {
	Page     int    `json:"page,default=1"`
	PageSize int    `json:"pageSize,default=20"`
	Name     string `json:"name,optional"`       // 按名称筛选
	TemplateId string `json:"templateId,optional"` // 按模板ID筛选
	Severity string `json:"severity,optional"`   // 按严重级别筛选
	Tag      string `json:"tag,optional"`        // 按标签筛选
	Enabled  *bool  `json:"enabled,optional"`    // 按状态筛选
}

type CustomPocListResp struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Total int         `json:"total"`
	List  []CustomPoc `json:"list"`
}

type CustomPocSaveReq struct {
	Id          string   `json:"id,optional"`
	Name        string   `json:"name"`
	TemplateId  string   `json:"templateId"`
	Severity    string   `json:"severity"`
	Tags        []string `json:"tags,optional"`
	Author      string   `json:"author,optional"`
	Description string   `json:"description,optional"`
	Content     string   `json:"content"`
	Enabled     bool     `json:"enabled"`
}

type CustomPocDeleteReq struct {
	Id string `json:"id"`
}

// CustomPocBatchImportReq 批量导入自定义POC请求
type CustomPocBatchImportReq struct {
	Pocs []CustomPocSaveReq `json:"pocs"` // POC列表
}

// CustomPocBatchImportResp 批量导入自定义POC响应
type CustomPocBatchImportResp struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Imported int    `json:"imported"` // 成功导入数量
	Failed   int    `json:"failed"`   // 失败数量
	Errors   []string `json:"errors"` // 错误信息列表
}

// CustomPocClearAllResp 清空所有自定义POC响应
type CustomPocClearAllResp struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Deleted int    `json:"deleted"` // 删除数量
}

// ==================== Nuclei默认模板 ====================
type NucleiTemplateListReq struct {
	Category string `json:"category,optional"` // 分类筛选
	Severity string `json:"severity,optional"` // 严重级别筛选
	Tag      string `json:"tag,optional"`      // 标签筛选
	Keyword  string `json:"keyword,optional"`  // 关键词搜索
	Page     int    `json:"page,default=1"`
	PageSize int    `json:"pageSize,default=50"`
	// 新增字段 - CVSS评分筛选和CVE搜索
	MinCvssScore float64 `json:"minCvssScore,optional"` // 最小CVSS评分筛选
	CveId        string  `json:"cveId,optional"`        // CVE编号搜索
}

type NucleiTemplate struct {
	Id          string   `json:"id"`          // 模板ID
	Name        string   `json:"name"`        // 模板名称
	Author      string   `json:"author"`      // 作者
	Severity    string   `json:"severity"`    // 严重级别
	Description string   `json:"description"` // 描述
	Tags        []string `json:"tags"`        // 标签
	Category    string   `json:"category"`    // 分类(目录名)
	FilePath    string   `json:"filePath"`    // 文件路径
	// 新增字段 - 漏洞知识库
	CvssScore   float64  `json:"cvssScore,omitempty"`   // CVSS评分
	CvssMetrics string   `json:"cvssMetrics,omitempty"` // CVSS向量
	CveIds      []string `json:"cveIds,omitempty"`      // CVE编号列表
	CweIds      []string `json:"cweIds,omitempty"`      // CWE编号列表
	References  []string `json:"references,omitempty"`  // 参考链接
	Remediation string   `json:"remediation,omitempty"` // 修复建议
}

type NucleiTemplateListResp struct {
	Code  int              `json:"code"`
	Msg   string           `json:"msg"`
	Total int              `json:"total"`
	List  []NucleiTemplate `json:"list"`
}

type NucleiTemplateCategoriesResp struct {
	Code       int               `json:"code"`
	Msg        string            `json:"msg"`
	Categories []string          `json:"categories"` // 分类列表
	Severities []string          `json:"severities"` // 严重级别列表
	Tags       []string          `json:"tags"`       // 常用标签列表
	Stats      map[string]int    `json:"stats"`      // 统计信息
}


type NucleiTemplateUpdateEnabledReq struct {
	TemplateIds []string `json:"templateIds"` // 模板ID列表
	Enabled     bool     `json:"enabled"`     // 启用/禁用
}


type NucleiTemplateDetailReq struct {
	TemplateId string `json:"templateId"` // 模板ID
}

type NucleiTemplateDetailResp struct {
	Code int                     `json:"code"`
	Msg  string                  `json:"msg"`
	Data *NucleiTemplateWithContent `json:"data"`
}

type NucleiTemplateWithContent struct {
	Id          string   `json:"id"`          // 模板ID
	Name        string   `json:"name"`        // 模板名称
	Author      string   `json:"author"`      // 作者
	Severity    string   `json:"severity"`    // 严重级别
	Description string   `json:"description"` // 描述
	Tags        []string `json:"tags"`        // 标签
	FilePath    string   `json:"filePath"`    // 文件路径
	Content     string   `json:"content"`     // YAML内容
	// 新增字段 - 漏洞知识库
	CvssScore   float64  `json:"cvssScore,omitempty"`   // CVSS评分
	CvssMetrics string   `json:"cvssMetrics,omitempty"` // CVSS向量
	CveIds      []string `json:"cveIds,omitempty"`      // CVE编号列表
	CweIds      []string `json:"cweIds,omitempty"`      // CWE编号列表
	References  []string `json:"references,omitempty"`  // 参考链接
	Remediation string   `json:"remediation,omitempty"` // 修复建议
}


// ==================== 指纹管理 ====================
type Fingerprint struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Website     string            `json:"website"`
	Icon        string            `json:"icon"`
	Description string            `json:"description"`
	Headers     map[string]string `json:"headers"`
	Cookies     map[string]string `json:"cookies"`
	HTML        []string          `json:"html"`
	Scripts     []string          `json:"scripts"`
	ScriptSrc   []string          `json:"scriptSrc"`
	JS          map[string]string `json:"js"`
	Meta        map[string]string `json:"meta"`
	CSS         []string          `json:"css"`
	URL         []string          `json:"url"`
	Dom         string            `json:"dom"`
	Rule        string            `json:"rule"`   // ARL格式规则
	Source      string            `json:"source"` // 来源: wappalyzer, arl, custom
	Implies     []string          `json:"implies"`
	Excludes    []string          `json:"excludes"`
	CPE         string            `json:"cpe"`
	IsBuiltin   bool              `json:"isBuiltin"`
	Enabled     bool              `json:"enabled"`
	CreateTime  string            `json:"createTime"`
	UpdateTime  string            `json:"updateTime"`
}

type FingerprintListReq struct {
	Page      int    `json:"page,default=1"`
	PageSize  int    `json:"pageSize,default=50"`
	Keyword   string `json:"keyword,optional"`
	Source    string `json:"source,optional"` // 来源筛选: arl, custom
	IsBuiltin *bool  `json:"isBuiltin,optional"`
	Enabled   *bool  `json:"enabled,optional"`
}

type FingerprintListResp struct {
	Code  int           `json:"code"`
	Msg   string        `json:"msg"`
	Total int           `json:"total"`
	List  []Fingerprint `json:"list"`
}

type FingerprintSaveReq struct {
	Id          string            `json:"id,optional"`
	Name        string            `json:"name"`
	Website     string            `json:"website,optional"`
	Icon        string            `json:"icon,optional"`
	Description string            `json:"description,optional"`
	Rule        string            `json:"rule,optional"`   // ARL格式规则
	Source      string            `json:"source,optional"` // 来源: custom, arl
	Headers     map[string]string `json:"headers,optional"`
	Cookies     map[string]string `json:"cookies,optional"`
	HTML        []string          `json:"html,optional"`
	Scripts     []string          `json:"scripts,optional"`
	Meta        map[string]string `json:"meta,optional"`
	CSS         []string          `json:"css,optional"`
	URL         []string          `json:"url,optional"`
	Implies     []string          `json:"implies,optional"`
	Excludes    []string          `json:"excludes,optional"`
	Enabled     bool              `json:"enabled"`
}

type FingerprintDeleteReq struct {
	Id string `json:"id"`
}

type FingerprintCategoriesResp struct {
	Code       int            `json:"code"`
	Msg        string         `json:"msg"`
	Categories []string       `json:"categories"`
	Stats      map[string]int64 `json:"stats"`
}

type FingerprintSyncReq struct {
	Force bool `json:"force"` // 强制重新同步
}

type FingerprintImportReq struct {
	Content   string `json:"content"`             // 文件内容
	Format    string `json:"format"`              // 格式: auto, arl-json, arl-yaml, finger-json, finger-yaml, wappalyzer
	IsBuiltin bool   `json:"isBuiltin,optional"` // 是否导入为内置指纹
}

type FingerprintImportResp struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Imported int    `json:"imported"` // 导入数量
	Skipped  int    `json:"skipped"`  // 跳过数量
}

// FingerprintImportFromFileReq 从文件/目录导入指纹
type FingerprintImportFromFileReq struct {
	Path string `json:"path"` // 文件或目录路径
}

// FingerprintClearCustomReq 清空自定义指纹请求
type FingerprintClearCustomReq struct {
	Source string `json:"source,optional"` // 可选：按来源清空，如 arl, arl-finger, custom；为空则清空所有自定义指纹
}

// FingerprintClearCustomResp 清空自定义指纹响应
type FingerprintClearCustomResp struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Deleted int    `json:"deleted"` // 删除数量
}

// FingerprintValidateReq 验证指纹请求
type FingerprintValidateReq struct {
	Id  string `json:"id,optional"`  // 指纹ID（验证已有指纹）
	Url string `json:"url"`          // 目标URL
}

// FingerprintValidateResp 验证指纹响应
type FingerprintValidateResp struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Matched bool   `json:"matched"` // 是否匹配
	Details string `json:"details"` // 匹配详情
}

// PocValidateReq 验证POC请求
type PocValidateReq struct {
	Id      string `json:"id,optional"`      // POC ID（验证已有POC）
	Url     string `json:"url"`              // 目标URL
	PocType string `json:"pocType,optional"` // POC类型: nuclei, custom (默认custom)
}

// PocValidateResp 验证POC响应
type PocValidateResp struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Matched  bool   `json:"matched"`  // 是否匹配/存在漏洞
	Severity string `json:"severity"` // 严重级别
	Details  string `json:"details"`  // 匹配详情
	TaskId   string `json:"taskId"`   // 任务ID（用于查询结果）
}

// PocBatchValidateReq 批量POC验证请求
type PocBatchValidateReq struct {
	Urls        []string `json:"urls"`                    // 目标URL列表
	PocType     string   `json:"pocType,optional"`        // POC类型: nuclei, custom, all (默认all)
	Severities  []string `json:"severities,optional"`     // 严重级别过滤
	Tags        []string `json:"tags,optional"`           // 标签过滤
	Timeout     int      `json:"timeout,optional"`        // 超时时间（秒，默认30）
	UseTemplate bool     `json:"useTemplate,optional"`    // 是否使用默认模板（默认true）
	UseCustom   bool     `json:"useCustom,optional"`      // 是否使用自定义POC（默认true）
	Concurrency int      `json:"concurrency,optional"`    // 并发数（默认10）
}

// PocValidationResult POC验证结果
type PocValidationResult struct {
	PocId      string   `json:"pocId"`      // POC ID
	PocName    string   `json:"pocName"`    // POC名称
	TemplateId string   `json:"templateId"` // 模板ID
	Severity   string   `json:"severity"`   // 严重级别
	Matched    bool     `json:"matched"`    // 是否匹配
	MatchedUrl string   `json:"matchedUrl"` // 匹配的URL
	Details    string   `json:"details"`    // 匹配详情
	Output     string   `json:"output"`     // 输出信息
	PocType    string   `json:"pocType"`    // POC类型: nuclei, custom
	Tags       []string `json:"tags"`       // 标签
}

// PocBatchValidateResp 批量POC验证响应
type PocBatchValidateResp struct {
	Code         int                            `json:"code"`
	Msg          string                         `json:"msg"`
	TotalUrls    int                            `json:"totalUrls"`    // 总URL数量
	TotalPocs    int                            `json:"totalPocs"`    // 总POC数量
	MatchedCount int                            `json:"matchedCount"` // 匹配数量
	Duration     string                         `json:"duration"`     // 耗时
	Results      []PocValidationResult          `json:"results"`      // 验证结果列表
	UrlStats     map[string]int                 `json:"urlStats"`     // 每个URL的匹配统计
	TaskId       string                         `json:"taskId"`       // 任务ID（用于查询结果）
	BatchId      string                         `json:"batchId"`      // 批次ID（用于查询结果）
}

// PocValidationResultQueryReq 查询POC验证结果请求
type PocValidationResultQueryReq struct {
	TaskId  string `json:"taskId,optional"`  // 任务ID
	BatchId string `json:"batchId,optional"` // 批次ID
}

// PocValidationResultQueryResp 查询POC验证结果响应
type PocValidationResultQueryResp struct {
	Code           int                   `json:"code"`
	Msg            string                `json:"msg"`
	Status         string                `json:"status"`         // 任务状态
	CompletedCount int                   `json:"completedCount"` // 已完成数量
	TotalCount     int                   `json:"totalCount"`     // 总数量
	Results        []PocValidationResult `json:"results"`        // 验证结果列表
	CreateTime     string                `json:"createTime"`     // 创建时间
	UpdateTime     string                `json:"updateTime"`     // 更新时间
}

// FingerprintBatchValidateReq 批量验证指纹请求
type FingerprintBatchValidateReq struct {
	Url   string `json:"url"`             // 目标URL
	Scope string `json:"scope,optional"`  // 范围: all, builtin, custom
}

// FingerprintBatchValidateResp 批量验证指纹响应
type FingerprintBatchValidateResp struct {
	Code         int                      `json:"code"`
	Msg          string                   `json:"msg"`
	MatchedCount int                      `json:"matchedCount"` // 匹配数量
	Duration     string                   `json:"duration"`     // 耗时
	Matched      []MatchedFingerprintInfo `json:"matched"`      // 匹配的指纹列表
}

// MatchedFingerprintInfo 匹配的指纹信息
type MatchedFingerprintInfo struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	IsBuiltin         bool   `json:"isBuiltin"`
	MatchedConditions string `json:"matchedConditions"` // 命中的条件
}

// ==================== HTTP服务映射 ====================
type HttpServiceMapping struct {
	Id          string `json:"id"`
	ServiceName string `json:"serviceName"` // 服务名称（小写）
	IsHttp      bool   `json:"isHttp"`      // 是否为HTTP服务
	Description string `json:"description"` // 描述
	Enabled     bool   `json:"enabled"`     // 是否启用
	CreateTime  string `json:"createTime"`
}

type HttpServiceMappingListReq struct {
	IsHttp  *bool  `json:"isHttp,optional"`  // 筛选：是否为HTTP服务
	Keyword string `json:"keyword,optional"` // 搜索：服务名称
}

type HttpServiceMappingListResp struct {
	Code int                  `json:"code"`
	Msg  string               `json:"msg"`
	List []HttpServiceMapping `json:"list"`
}

type HttpServiceMappingSaveReq struct {
	Id          string `json:"id,optional"`
	ServiceName string `json:"serviceName"`
	IsHttp      bool   `json:"isHttp"`
	Description string `json:"description,optional"`
	Enabled     bool   `json:"enabled"`
}

type HttpServiceMappingDeleteReq struct {
	Id string `json:"id"`
}


// ==================== 报告管理 ====================
type ReportDetailReq struct {
	TaskId string `json:"taskId"`
}

type ReportAsset struct {
	Authority  string   `json:"authority"`
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	Service    string   `json:"service"`
	Title      string   `json:"title"`
	App        []string `json:"app"`
	HttpStatus string   `json:"httpStatus"`
	Server     string   `json:"server"`
	IconHash   string   `json:"iconHash"`
	Screenshot string   `json:"screenshot"`
	CreateTime string   `json:"createTime"`
}

type ReportVul struct {
	Authority  string `json:"authority"`
	Url        string `json:"url"`
	PocFile    string `json:"pocFile"`
	Severity   string `json:"severity"`
	Result     string `json:"result"`
	CreateTime string `json:"createTime"`
}

type ReportData struct {
	TaskId      string         `json:"taskId"`
	TaskName    string         `json:"taskName"`
	Target      string         `json:"target"`
	Status      string         `json:"status"`
	CreateTime  string         `json:"createTime"`
	AssetCount  int            `json:"assetCount"`
	VulCount    int            `json:"vulCount"`
	Assets      []ReportAsset  `json:"assets"`
	Vuls        []ReportVul    `json:"vuls"`
	TopPorts    []StatItem     `json:"topPorts"`
	TopServices []StatItem     `json:"topServices"`
	TopApps     []StatItem     `json:"topApps"`
	VulStats    map[string]int `json:"vulStats"`
}

type ReportDetailResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data *ReportData `json:"data"`
}

type ReportExportReq struct {
	TaskId string `json:"taskId"`
	Format string `json:"format,optional"` // excel, pdf (默认excel)
}
