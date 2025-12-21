package xerr

// 错误码定义
const (
	OK          = 0
	ParamError  = 400
	Unauthorized = 401
	Forbidden   = 403
	NotFound    = 404
	ServerError = 500

	// 业务错误码 10000+
	UserNotFound       = 10001
	UserPasswordError  = 10002
	UserDisabled       = 10003
	TaskNotFound       = 10101
	ProfileNotFound    = 10102
	TaskStatusError    = 10103
	WorkspaceNotFound  = 10201
	AssetNotFound      = 10301
	VulNotFound        = 10401
	FingerprintNotFound = 10501
	PocNotFound        = 10601
)

var codeMsg = map[int]string{
	OK:                  "success",
	ParamError:          "参数错误",
	Unauthorized:        "未授权",
	Forbidden:           "禁止访问",
	NotFound:            "资源不存在",
	ServerError:         "服务器错误",
	UserNotFound:        "用户不存在",
	UserPasswordError:   "用户名或密码错误",
	UserDisabled:        "用户已禁用",
	TaskNotFound:        "任务不存在",
	ProfileNotFound:     "任务配置不存在",
	TaskStatusError:     "任务状态不允许此操作",
	WorkspaceNotFound:   "工作空间不存在",
	AssetNotFound:       "资产不存在",
	VulNotFound:         "漏洞不存在",
	FingerprintNotFound: "指纹不存在",
	PocNotFound:         "POC不存在",
}

// GetMsg 获取错误信息
func GetMsg(code int) string {
	if msg, ok := codeMsg[code]; ok {
		return msg
	}
	return "未知错误"
}
