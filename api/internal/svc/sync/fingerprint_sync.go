package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cscan/model"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/zeromicro/go-zero/core/logx"
)

// FingerprintSyncService 指纹同步服务
type FingerprintSyncService struct {
	model *model.FingerprintModel
}

// NewFingerprintSyncService 创建指纹同步服务
func NewFingerprintSyncService(model *model.FingerprintModel) *FingerprintSyncService {
	return &FingerprintSyncService{model: model}
}

// SyncWappalyzerFingerprints 同步Wappalyzer指纹到数据库
func (s *FingerprintSyncService) SyncWappalyzerFingerprints(ctx context.Context) {
	count, _ := s.model.Count(ctx, map[string]interface{}{"is_builtin": true})
	if count > 0 {
		logx.Infof("[FingerprintSync] Found %d builtin fingerprints, skipping sync", count)
		return
	}

	logx.Info("[FingerprintSync] Starting sync of Wappalyzer fingerprints...")
	startTime := time.Now()

	fingerprintsData := wappalyzer.GetFingerprints()
	if fingerprintsData == "" {
		logx.Info("[FingerprintSync] No fingerprints data available")
		return
	}

	var fingerprints wappalyzer.Fingerprints
	if err := json.Unmarshal([]byte(fingerprintsData), &fingerprints); err != nil {
		logx.Errorf("[FingerprintSync] Failed to parse fingerprints: %v", err)
		return
	}

	categoryNames := getCategoryNames()

	totalCount := 0
	for name, fp := range fingerprints.Apps {
		categoryName := "unknown"
		if len(fp.Cats) > 0 {
			if catName, ok := categoryNames[fp.Cats[0]]; ok {
				categoryName = catName
			} else {
				categoryName = fmt.Sprintf("category-%d", fp.Cats[0])
			}
		}

		metaMap := make(map[string]string)
		for k, v := range fp.Meta {
			if len(v) > 0 {
				metaMap[k] = strings.Join(v, " | ")
			}
		}

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

		if err := s.model.Upsert(ctx, doc); err != nil {
			logx.Errorf("[FingerprintSync] Failed to upsert %s: %v", name, err)
			continue
		}
		totalCount++
	}

	duration := time.Since(startTime)
	logx.Infof("[FingerprintSync] Completed: %d fingerprints synced in %v", totalCount, duration)
}

// getCategoryNames 获取Wappalyzer分类映射
func getCategoryNames() map[int]string {
	return map[int]string{
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
		67: "Cookie compliance", 68: "Accessibility", 69: "Social login",
		70: "SSL/TLS certificate authorities", 71: "Affiliate programs",
		72: "Appointment scheduling", 73: "Surveys", 74: "A/B Testing",
		75: "Email", 76: "Personalisation", 77: "Retargeting", 78: "RUM",
		79: "Geolocation", 80: "WordPress themes", 81: "Shopify themes",
		82: "WordPress plugins", 83: "Shopify apps", 84: "Drupal themes",
		85: "Browser fingerprinting", 86: "Loyalty & rewards", 87: "Feature management",
		88: "Segmentation", 89: "Hosting", 90: "Translation", 91: "Reviews",
		92: "Buy now pay later", 93: "Performance", 94: "Reservations & delivery",
		95: "Referral marketing", 96: "Digital asset management", 97: "Content curation",
		98: "Customer data platform", 99: "Cart abandonment", 100: "Shipping carriers",
		101: "Fulfilment", 102: "Returns", 103: "Cross border ecommerce",
	}
}
