package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type ContextKey string

const (
	UserIdKey      ContextKey = "userId"
	UsernameKey    ContextKey = "username"
	RoleKey        ContextKey = "role"
	WorkspaceIdKey ContextKey = "workspaceId"
)

type AuthMiddleware struct {
	AccessSecret string
}

func NewAuthMiddleware(accessSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		AccessSecret: accessSecret,
	}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从Header获取Token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			unauthorized(w, "未提供认证信息")
			return
		}

		// 解析Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			unauthorized(w, "认证格式错误")
			return
		}

		// 验证Token
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(m.AccessSecret), nil
		})
		if err != nil || !token.Valid {
			unauthorized(w, "Token无效或已过期")
			return
		}

		// 提取Claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			unauthorized(w, "Token解析失败")
			return
		}

		// 将用户信息存入Context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIdKey, claims["userId"])
		ctx = context.WithValue(ctx, UsernameKey, claims["username"])
		ctx = context.WithValue(ctx, RoleKey, claims["role"])

		// 从Header获取当前工作空间
		workspaceId := r.Header.Get("X-Workspace-Id")
		ctx = context.WithValue(ctx, WorkspaceIdKey, workspaceId)

		next(w, r.WithContext(ctx))
	}
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 401,
		"msg":  msg,
	})
}

// GetUserId 从Context获取用户ID
func GetUserId(ctx context.Context) string {
	if v := ctx.Value(UserIdKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetUsername 从Context获取用户名
func GetUsername(ctx context.Context) string {
	if v := ctx.Value(UsernameKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetRole 从Context获取角色
func GetRole(ctx context.Context) string {
	if v := ctx.Value(RoleKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetWorkspaceId 从Context获取工作空间ID
func GetWorkspaceId(ctx context.Context) string {
	if v := ctx.Value(WorkspaceIdKey); v != nil {
		return v.(string)
	}
	return ""
}
