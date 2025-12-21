package common

import (
	"context"
	"strings"

	"cscan/api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// InjectPocConfig 注入POC模板ID到任务配置（不存储完整内容，避免文档过大）
func InjectPocConfig(ctx context.Context, svcCtx *svc.ServiceContext, taskConfig map[string]interface{}, logger logx.Logger) map[string]interface{} {
	pocscan, ok := taskConfig["pocscan"].(map[string]interface{})
	if !ok || pocscan == nil {
		return taskConfig
	}

	// 检查是否启用POC扫描和Nuclei
	enable, _ := pocscan["enable"].(bool)
	useNuclei, _ := pocscan["useNuclei"].(bool)
	if !enable || !useNuclei {
		return taskConfig
	}

	// 检查是否启用自动扫描模式
	autoScan, _ := pocscan["autoScan"].(bool)
	automaticScan, _ := pocscan["automaticScan"].(bool)

	// 如果启用了自动扫描，不预先注入模板ID，让Worker根据资产指纹动态获取
	if autoScan || automaticScan {
		logger.Infof("Auto-scan enabled (autoScan=%v, automaticScan=%v), skipping template ID injection", autoScan, automaticScan)

		// 只注入标签映射（用于自定义标签映射模式）
		if autoScan {
			tagMappings, err := svcCtx.TagMappingModel.FindEnabled(ctx)
			if err == nil && len(tagMappings) > 0 {
				mappings := make(map[string][]string)
				for _, tm := range tagMappings {
					mappings[tm.AppName] = tm.NucleiTags
				}
				pocscan["tagMappings"] = mappings
				logger.Infof("Injected %d tag mappings for auto-scan", len(mappings))
			}
		}

		taskConfig["pocscan"] = pocscan
		return taskConfig
	}

	customPocOnly, _ := pocscan["customPocOnly"].(bool)
	var nucleiTemplateIds []string
	var customPocIds []string

	if customPocOnly {
		// 只使用自定义POC - 存储ID列表
		customPocs, err := svcCtx.CustomPocModel.FindEnabled(ctx)
		if err == nil && len(customPocs) > 0 {
			for _, poc := range customPocs {
				customPocIds = append(customPocIds, poc.Id.Hex())
			}
			logger.Infof("Injected %d custom POC IDs (CustomPocOnly mode)", len(customPocIds))
		}
	} else {
		// 从数据库获取默认模板ID（根据严重级别筛选）
		severityStr, _ := pocscan["severity"].(string)
		if severityStr != "" {
			severities := strings.Split(severityStr, ",")
			nucleiTemplates, err := svcCtx.NucleiTemplateModel.FindBySeverity(ctx, severities)
			if err == nil && len(nucleiTemplates) > 0 {
				for _, t := range nucleiTemplates {
					nucleiTemplateIds = append(nucleiTemplateIds, t.Id.Hex())
				}
				logger.Infof("Injected %d nuclei template IDs (severity: %s)", len(nucleiTemplateIds), severityStr)
			}
		}

		// 添加自定义POC ID
		customPocs, err := svcCtx.CustomPocModel.FindEnabled(ctx)
		if err == nil && len(customPocs) > 0 {
			for _, poc := range customPocs {
				customPocIds = append(customPocIds, poc.Id.Hex())
			}
			logger.Infof("Added %d custom POC IDs", len(customPocIds))
		}
	}

	// 存储ID列表而不是完整内容
	if len(nucleiTemplateIds) > 0 {
		pocscan["nucleiTemplateIds"] = nucleiTemplateIds
	}
	if len(customPocIds) > 0 {
		pocscan["customPocIds"] = customPocIds
	}

	taskConfig["pocscan"] = pocscan
	return taskConfig
}
