package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Response 结构体定义
type Response struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	MsgCode string      `json:"msgCode"`
	Data    interface{} `json:"data,omitempty"`
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponseData 登录响应数据
type LoginResponseData struct {
	Token string `json:"token"`
}

// User 模拟用户数据
type User struct {
	ID       string
	Username string
	Password string
}

// UpdateNameRequest 修改用户名请求结构体
type UpdateNameRequest struct {
	UserID  string `json:"userId"`
	NewName string `json:"newName"`
}

// 模拟用户数据库
var mockUsers = map[string]User{
	"testuser": {ID: "1", Username: "testuser", Password: "testpass"},
}

// 模拟 token 存储
var validTokens = map[string]string{} // token -> userID
var tokenMutex = sync.Mutex{}

// 默认成功响应
var defaultSuccessResponse = Response{
	Code:    200,
	Msg:     "success",
	MsgCode: "0",
}

// JWT 密钥
const jwtSecret = "test-secret"

// 全局 token
var testToken string

// AssertionLogger 封装 zap 日志
type AssertionLogger struct {
	logger *zap.Logger
}

// NewAssertionLogger 初始化 zap 日志
func NewAssertionLogger() (*AssertionLogger, error) {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			TimeKey:      "time",
			CallerKey:    "caller",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	logger, err := cfg.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize zap logger")
	}
	return &AssertionLogger{logger: logger}, nil
}

// LogAssertion 记录断言失败结果
func (al *AssertionLogger) LogAssertion(t *testing.T, field, testName string, actual, expected interface{}, passed bool) {
	if !passed {
		al.logger.Error(
			"Assertion failed: "+field+"预期"+interfaceToString(expected)+"，实际返回"+interfaceToString(actual)+"，"+field+"值断言失败",
			zap.String("test_name", testName),
			zap.String("field", field),
			zap.Any("expected", expected),
			zap.Any("actual", actual),
		)
		t.Fail() // 标记测试失败
	}
}

// interfaceToString 将 interface{} 转换为字符串表示
func interfaceToString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return `"` + v + `"`
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// generateJWT 生成 JWT token
func generateJWT(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	tokenMutex.Lock()
	validTokens[tokenString] = userID
	tokenMutex.Unlock()
	return tokenString, nil
}

// validateJWT 验证 JWT token
func validateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	tokenMutex.Lock()
	userID, exists := validTokens[tokenString]
	tokenMutex.Unlock()
	if !exists {
		return "", fmt.Errorf("token not found")
	}
	return userID, nil
}

// 登录接口
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := Response{Code: 400, Msg: "Invalid request", MsgCode: "INVALID_REQUEST"}
		json.NewEncoder(w).Encode(response)
		return
	}

	user, exists := mockUsers[req.Username]
	if !exists || user.Password != req.Password {
		response := Response{Code: 401, Msg: "Invalid credentials", MsgCode: "UNAUTHORIZED"}
		json.NewEncoder(w).Encode(response)
		return
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		response := Response{Code: 500, Msg: "Failed to generate token", MsgCode: "INTERNAL_ERROR"}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := Response{
		Code:    200,
		Msg:     "success",
		MsgCode: "0",
		Data:    LoginResponseData{Token: token},
	}
	json.NewEncoder(w).Encode(response)
}

// 用户详情接口
func userDetailsHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		response := Response{Code: 401, Msg: "Missing or invalid token", MsgCode: "UNAUTHORIZED"}
		json.NewEncoder(w).Encode(response)
		return
	}
	token := authHeader[7:]
	userID, err := validateJWT(token)
	if err != nil {
		response := Response{Code: 401, Msg: "Invalid token", MsgCode: "UNAUTHORIZED"}
		json.NewEncoder(w).Encode(response)
		return
	}

	queryUserID := r.URL.Query().Get("user_id")
	if queryUserID == "" {
		response := Response{Code: 400, Msg: "User ID missing", MsgCode: "INVALID_REQUEST"}
		json.NewEncoder(w).Encode(response)
		return
	}

	if queryUserID != userID {
		response := Response{Code: 403, Msg: "Forbidden: user ID mismatch", MsgCode: "FORBIDDEN"}
		json.NewEncoder(w).Encode(response)
		return
	}

	if _, exists := mockUsers[queryUserID]; !exists {
		response := Response{Code: 404, Msg: "User not found", MsgCode: "NOT_FOUND"}
		json.NewEncoder(w).Encode(response)
		return
	}

	json.NewEncoder(w).Encode(defaultSuccessResponse)
}

// 修改用户名接口
func updateNameHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		response := Response{Code: 401, Msg: "Missing or invalid token", MsgCode: "UNAUTHORIZED"}
		json.NewEncoder(w).Encode(response)
		return
	}
	token := authHeader[7:]
	userID, err := validateJWT(token)
	if err != nil {
		response := Response{Code: 401, Msg: "Invalid token", MsgCode: "UNAUTHORIZED"}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req UpdateNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := Response{Code: 400, Msg: "Invalid request", MsgCode: "INVALID_REQUEST"}
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.UserID == "" {
		response := Response{Code: 400, Msg: "User ID missing", MsgCode: "INVALID_REQUEST"}
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.UserID != userID {
		response := Response{Code: 403, Msg: "Forbidden: user ID mismatch", MsgCode: "FORBIDDEN"}
		json.NewEncoder(w).Encode(response)
		return
	}

	if user, exists := mockUsers[req.UserID]; exists {
		user.Username = req.NewName
		mockUsers[req.UserID] = user
		json.NewEncoder(w).Encode(defaultSuccessResponse)
		return
	}

	response := Response{Code: 404, Msg: "User not found", MsgCode: "NOT_FOUND"}
	json.NewEncoder(w).Encode(response)
}

// 退出登录接口
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		response := Response{Code: 401, Msg: "Missing or invalid token", MsgCode: "UNAUTHORIZED"}
		json.NewEncoder(w).Encode(response)
		return
	}
	token := authHeader[7:]
	tokenMutex.Lock()
	if _, exists := validTokens[token]; exists {
		delete(validTokens, token)
		tokenMutex.Unlock()
		json.NewEncoder(w).Encode(defaultSuccessResponse)
		return
	}
	tokenMutex.Unlock()
	response := Response{Code: 401, Msg: "Invalid token", MsgCode: "UNAUTHORIZED"}
	json.NewEncoder(w).Encode(response)
}

// 获取 token 的辅助函数
func getToken(t *testing.T, username, password string) string {
	body, _ := json.Marshal(LoginRequest{Username: username, Password: password})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	loginHandler(w, req)

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}
	if resp.Code != 200 {
		t.Fatalf("Login failed: %v", resp)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Invalid login response data")
	}
	token, ok := data["token"].(string)
	if !ok {
		t.Fatalf("Token not found in login response")
	}
	return token
}

// 退出登录的辅助函数
func logout(t *testing.T, token string) {
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	logoutHandler(w, req)

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode logout response: %v", err)
	}
	if resp.Code != 200 {
		t.Fatalf("Logout failed: %v", resp)
	}
}

// TestMain 设置前置和后置处理器
func TestMain(m *testing.M) {
	// 前置处理器：获取 token
	testToken = getToken(&testing.T{}, "testuser", "testpass")

	// 运行测试
	code := m.Run()

	// 后置处理器：退出登录
	logout(&testing.T{}, testToken)

	os.Exit(code)
}

// 测试登录接口
func TestLoginAPI(t *testing.T) {
	// 初始化日志
	assertionLogger, err := NewAssertionLogger()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer assertionLogger.logger.Sync()

	// 测试成功场景
	t.Run("Login success", func(t *testing.T) {
		t.Parallel() // 并行运行
		body, _ := json.Marshal(LoginRequest{Username: "testuser", Password: "testpass"})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		loginHandler(w, req)

		var actual Response
		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// 逐字段断言，仅失败时记录日志
		assertionLogger.LogAssertion(t, "code", "Login success", actual.Code, defaultSuccessResponse.Code,
			assert.True(t, actual.Code == defaultSuccessResponse.Code))
		assertionLogger.LogAssertion(t, "msg", "Login success", actual.Msg, defaultSuccessResponse.Msg,
			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
		assertionLogger.LogAssertion(t, "msgCode", "Login success", actual.MsgCode, defaultSuccessResponse.MsgCode,
			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
	})
}

// 测试用户详情接口
func TestUserDetailsAPI(t *testing.T) {
	// 初始化日志
	assertionLogger, err := NewAssertionLogger()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer assertionLogger.logger.Sync()

	// 测试成功场景
	t.Run("User details success", func(t *testing.T) {
		t.Parallel() // 并行运行
		req := httptest.NewRequest(http.MethodGet, "/user?user_id=1", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()

		userDetailsHandler(w, req)

		var actual Response
		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// 逐字段断言，仅失败时记录日志
		assertionLogger.LogAssertion(t, "code", "User details success", actual.Code, defaultSuccessResponse.Code,
			assert.True(t, actual.Code == defaultSuccessResponse.Code))
		assertionLogger.LogAssertion(t, "msg", "User details success", actual.Msg, defaultSuccessResponse.Msg,
			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
		assertionLogger.LogAssertion(t, "msgCode", "User details success", actual.MsgCode, defaultSuccessResponse.MsgCode,
			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
	})
}

// 测试修改用户名接口
func TestUpdateNameAPI(t *testing.T) {
	// 初始化日志
	assertionLogger, err := NewAssertionLogger()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer assertionLogger.logger.Sync()

	// 测试成功场景
	t.Run("Update name success", func(t *testing.T) {
		t.Parallel() // 并行运行
		body, _ := json.Marshal(UpdateNameRequest{UserID: "1", NewName: "newuser"})
		req := httptest.NewRequest(http.MethodPut, "/user/name", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()

		updateNameHandler(w, req)

		var actual Response
		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// 逐字段断言，仅失败时记录日志
		assertionLogger.LogAssertion(t, "code", "Update name success", actual.Code, defaultSuccessResponse.Code,
			assert.True(t, actual.Code == defaultSuccessResponse.Code))
		assertionLogger.LogAssertion(t, "msg", "Update name success", actual.Msg, defaultSuccessResponse.Msg,
			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
		assertionLogger.LogAssertion(t, "msgCode", "Update name success", actual.MsgCode, defaultSuccessResponse.MsgCode,
			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
	})
}
