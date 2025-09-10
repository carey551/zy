package demoreport

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/pkg/errors"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/zap"
// 	"go.uber.org/zap/zapcore"
// )

// // Response 结构体定义
// type Response struct {
// 	Code    int    `json:"code"`
// 	Msg     string `json:"msg"`
// 	MsgCode string `json:"msgCode"`
// }

// // LoginRequest 登录请求结构体
// type LoginRequest struct {
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// }

// // User 模拟用户数据
// type User struct {
// 	ID       string
// 	Username string
// 	Password string
// }

// // UpdateNameRequest 修改用户名请求结构体
// type UpdateNameRequest struct {
// 	UserID  string `json:"userId"`
// 	NewName string `json:"newName"`
// }

// // 模拟用户数据库
// var mockUsers = map[string]User{
// 	"testuser": {ID: "1", Username: "testuser", Password: "testpass"},
// }

// // 默认成功响应
// var defaultSuccessResponse = Response{
// 	Code:    200,
// 	Msg:     "success",
// 	MsgCode: "0",
// }

// // AssertionLogger 封装 zap 日志
// type AssertionLogger struct {
// 	logger *zap.Logger
// }

// // NewAssertionLogger 初始化 zap 日志
// func NewAssertionLogger() (*AssertionLogger, error) {
// 	cfg := zap.Config{
// 		Encoding:         "json",
// 		Level:            zap.NewAtomicLevelAt(zapcore.ErrorLevel), // 仅记录 Error 级别日志
// 		OutputPaths:      []string{"stderr"},
// 		ErrorOutputPaths: []string{"stderr"},
// 		EncoderConfig: zapcore.EncoderConfig{
// 			MessageKey:   "message",
// 			LevelKey:     "level",
// 			TimeKey:      "time",
// 			CallerKey:    "caller",
// 			EncodeLevel:  zapcore.CapitalLevelEncoder,
// 			EncodeTime:   zapcore.ISO8601TimeEncoder,
// 			EncodeCaller: zapcore.ShortCallerEncoder,
// 		},
// 	}
// 	logger, err := cfg.Build()
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to initialize zap logger")
// 	}
// 	return &AssertionLogger{logger: logger}, nil
// }

// // LogAssertion 记录断言失败结果
// func (al *AssertionLogger) LogAssertion(t *testing.T, field, testName string, actual, expected interface{}, passed bool) {
// 	if !passed {
// 		al.logger.Error(
// 			"Assertion failed: "+field+"预期"+interfaceToString(expected)+"，实际返回"+interfaceToString(actual)+"，"+field+"值断言失败",
// 			zap.String("test_name", testName),
// 			zap.String("field", field),
// 			zap.Any("expected", expected),
// 			zap.Any("actual", actual),
// 		)
// 		t.Fail() // 标记测试失败
// 	}
// }

// // interfaceToString 将 interface{} 转换为字符串表示
// func interfaceToString(v interface{}) string {
// 	switch v := v.(type) {
// 	case string:
// 		return `"` + v + `"`
// 	case int:
// 		return fmt.Sprintf("%d", v)
// 	default:
// 		return fmt.Sprintf("%v", v)
// 	}
// }

// // 登录接口
// func loginHandler(w http.ResponseWriter, r *http.Request) {
// 	var req LoginRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		response := Response{Code: 400, Msg: "Invalid request", MsgCode: "INVALID_REQUEST"}
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	user, exists := mockUsers[req.Username]
// 	if !exists || user.Password != req.Password {
// 		response := Response{Code: 401, Msg: "Invalid credentials", MsgCode: "UNAUTHORIZED"}
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(defaultSuccessResponse)
// }

// // 用户详情接口
// func userDetailsHandler(w http.ResponseWriter, r *http.Request) {
// 	userID := r.URL.Query().Get("user_id")
// 	if userID == "" {
// 		response := Response{Code: 400, Msg: "User ID missing", MsgCode: "INVALID_REQUEST"}
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	if _, exists := mockUsers[userID]; !exists {
// 		response := Response{Code: 404, Msg: "User not found", MsgCode: "NOT_FOUND"}
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(defaultSuccessResponse)
// }

// // 修改用户名接口
// func updateNameHandler(w http.ResponseWriter, r *http.Request) {
// 	var req UpdateNameRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		response := Response{Code: 400, Msg: "Invalid request", MsgCode: "INVALID_REQUEST"}
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	if req.UserID == "" {
// 		response := Response{Code: 400, Msg: "User ID missing", MsgCode: "INVALID_REQUEST"}
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	if user, exists := mockUsers[req.UserID]; exists {
// 		user.Username = req.NewName
// 		mockUsers[req.UserID] = user
// 		json.NewEncoder(w).Encode(defaultSuccessResponse)
// 		return
// 	}

// 	response := Response{Code: 404, Msg: "User not found", MsgCode: "NOT_FOUND"}
// 	json.NewEncoder(w).Encode(response)
// }

// // 测试登录接口
// func TestLoginAPI(t *testing.T) {
// 	// 初始化日志
// 	assertionLogger, err := NewAssertionLogger()
// 	if err != nil {
// 		t.Fatalf("Failed to initialize logger: %v", err)
// 	}
// 	defer assertionLogger.logger.Sync()

// 	// 测试成功场景
// 	t.Run("Login success", func(t *testing.T) {
// 		body, _ := json.Marshal(LoginRequest{Username: "testuser", Password: "testpass"})
// 		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
// 		w := httptest.NewRecorder()

// 		loginHandler(w, req)

// 		var actual Response
// 		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
// 			t.Fatalf("Failed to decode response: %v", err)
// 		}

// 		// 逐字段断言，仅失败时记录日志
// 		assertionLogger.LogAssertion(t, "code", "Login success", actual.Code, defaultSuccessResponse.Code,
// 			assert.True(t, actual.Code == defaultSuccessResponse.Code))
// 		assertionLogger.LogAssertion(t, "msg", "Login success", actual.Msg, defaultSuccessResponse.Msg,
// 			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
// 		assertionLogger.LogAssertion(t, "msgCode", "Login success", actual.MsgCode, defaultSuccessResponse.MsgCode,
// 			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
// 	})

// 	// 测试失败场景 - 错误密码
// 	t.Run("Login failure - wrong password", func(t *testing.T) {
// 		body, _ := json.Marshal(LoginRequest{Username: "testuser", Password: "wrongpass"})
// 		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
// 		w := httptest.NewRecorder()

// 		loginHandler(w, req)

// 		var actual Response
// 		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
// 			t.Fatalf("Failed to decode response: %v", err)
// 		}

// 		// 逐字段断言，仅失败时记录日志
// 		assertionLogger.LogAssertion(t, "code", "Login failure - wrong password", actual.Code, defaultSuccessResponse.Code,
// 			assert.True(t, actual.Code == defaultSuccessResponse.Code))
// 		assertionLogger.LogAssertion(t, "msg", "Login failure - wrong password", actual.Msg, defaultSuccessResponse.Msg,
// 			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
// 		assertionLogger.LogAssertion(t, "msgCode", "Login failure - wrong password", actual.MsgCode, defaultSuccessResponse.MsgCode,
// 			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
// 	})
// }

// // 测试用户详情接口
// func TestUserDetailsAPI(t *testing.T) {
// 	// 初始化日志
// 	assertionLogger, err := NewAssertionLogger()
// 	if err != nil {
// 		t.Fatalf("Failed to initialize logger: %v", err)
// 	}
// 	defer assertionLogger.logger.Sync()

// 	// 测试成功场景
// 	t.Run("User details success", func(t *testing.T) {
// 		req := httptest.NewRequest(http.MethodGet, "/user?user_id=1", nil)
// 		w := httptest.NewRecorder()

// 		userDetailsHandler(w, req)

// 		var actual Response
// 		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
// 			t.Fatalf("Failed to decode response: %v", err)
// 		}

// 		// 逐字段断言，仅失败时记录日志
// 		assertionLogger.LogAssertion(t, "code", "User details success", actual.Code, defaultSuccessResponse.Code,
// 			assert.True(t, actual.Code == defaultSuccessResponse.Code))
// 		assertionLogger.LogAssertion(t, "msg", "User details success", actual.Msg, defaultSuccessResponse.Msg,
// 			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
// 		assertionLogger.LogAssertion(t, "msgCode", "User details success", actual.MsgCode, defaultSuccessResponse.MsgCode,
// 			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
// 	})

// 	// 测试失败场景 - 无 user_id
// 	t.Run("User details failure - no user_id", func(t *testing.T) {
// 		req := httptest.NewRequest(http.MethodGet, "/user", nil)
// 		w := httptest.NewRecorder()

// 		userDetailsHandler(w, req)

// 		var actual Response
// 		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
// 			t.Fatalf("Failed to decode response: %v", err)
// 		}

// 		// 逐字段断言，仅失败时记录日志
// 		assertionLogger.LogAssertion(t, "code", "User details failure - no user_id", actual.Code, defaultSuccessResponse.Code,
// 			assert.True(t, actual.Code == defaultSuccessResponse.Code))
// 		assertionLogger.LogAssertion(t, "msg", "User details failure - no user_id", actual.Msg, defaultSuccessResponse.Msg,
// 			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
// 		assertionLogger.LogAssertion(t, "msgCode", "User details failure - no user_id", actual.MsgCode, defaultSuccessResponse.MsgCode,
// 			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
// 	})

// 	// 测试失败场景 - 用户不存在
// 	t.Run("User details failure - user not found", func(t *testing.T) {
// 		req := httptest.NewRequest(http.MethodGet, "/user?user_id=999", nil)
// 		w := httptest.NewRecorder()

// 		userDetailsHandler(w, req)

// 		var actual Response
// 		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
// 			t.Fatalf("Failed to decode response: %v", err)
// 		}

// 		// 逐字段断言，仅失败时记录日志
// 		assertionLogger.LogAssertion(t, "code", "User details failure - user not found", actual.Code, defaultSuccessResponse.Code,
// 			assert.True(t, actual.Code == defaultSuccessResponse.Code))
// 		assertionLogger.LogAssertion(t, "msg", "User details failure - user not found", actual.Msg, defaultSuccessResponse.Msg,
// 			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
// 		assertionLogger.LogAssertion(t, "msgCode", "User details failure - user not found", actual.MsgCode, defaultSuccessResponse.MsgCode,
// 			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
// 	})
// }

// // 测试修改用户名接口
// func TestUpdateNameAPI(t *testing.T) {
// 	// 初始化日志
// 	assertionLogger, err := NewAssertionLogger()
// 	if err != nil {
// 		t.Fatalf("Failed to initialize logger: %v", err)
// 	}
// 	defer assertionLogger.logger.Sync()

// 	// 测试成功场景
// 	t.Run("Update name success", func(t *testing.T) {
// 		body, _ := json.Marshal(UpdateNameRequest{UserID: "1", NewName: "newuser"})
// 		req := httptest.NewRequest(http.MethodPut, "/user/name", bytes.NewReader(body))
// 		w := httptest.NewRecorder()

// 		updateNameHandler(w, req)

// 		var actual Response
// 		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
// 			t.Fatalf("Failed to decode response: %v", err)
// 		}

// 		// 逐字段断言，仅失败时记录日志
// 		assertionLogger.LogAssertion(t, "code", "Update name success", actual.Code, defaultSuccessResponse.Code,
// 			assert.True(t, actual.Code == defaultSuccessResponse.Code))
// 		assertionLogger.LogAssertion(t, "msg", "Update name success", actual.Msg, defaultSuccessResponse.Msg,
// 			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
// 		assertionLogger.LogAssertion(t, "msgCode", "Update name success", actual.MsgCode, defaultSuccessResponse.MsgCode,
// 			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
// 	})

// 	// 测试失败场景 - 无 user_id
// 	t.Run("Update name failure - no user_id", func(t *testing.T) {
// 		body, _ := json.Marshal(UpdateNameRequest{NewName: "newuser"})
// 		req := httptest.NewRequest(http.MethodPut, "/user/name", bytes.NewReader(body))
// 		w := httptest.NewRecorder()

// 		updateNameHandler(w, req)

// 		var actual Response
// 		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
// 			t.Fatalf("Failed to decode response: %v", err)
// 		}

// 		// 逐字段断言，仅失败时记录日志
// 		assertionLogger.LogAssertion(t, "code", "Update name failure - no user_id", actual.Code, defaultSuccessResponse.Code,
// 			assert.True(t, actual.Code == defaultSuccessResponse.Code))
// 		assertionLogger.LogAssertion(t, "msg", "Update name failure - no user_id", actual.Msg, defaultSuccessResponse.Msg,
// 			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
// 		assertionLogger.LogAssertion(t, "msgCode", "Update name failure - no user_id", actual.MsgCode, defaultSuccessResponse.MsgCode,
// 			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
// 	})

// 	// 测试失败场景 - 用户不存在
// 	t.Run("Update name failure - user not found", func(t *testing.T) {
// 		body, _ := json.Marshal(UpdateNameRequest{UserID: "999", NewName: "newuser"})
// 		req := httptest.NewRequest(http.MethodPut, "/user/name", bytes.NewReader(body))
// 		w := httptest.NewRecorder()

// 		updateNameHandler(w, req)

// 		var actual Response
// 		if err := json.NewDecoder(w.Body).Decode(&actual); err != nil {
// 			t.Fatalf("Failed to decode response: %v", err)
// 		}

// 		// 逐字段断言，仅失败时记录日志
// 		assertionLogger.LogAssertion(t, "code", "Update name failure - user not found", actual.Code, defaultSuccessResponse.Code,
// 			assert.True(t, actual.Code == defaultSuccessResponse.Code))
// 		assertionLogger.LogAssertion(t, "msg", "Update name failure - user not found", actual.Msg, defaultSuccessResponse.Msg,
// 			assert.True(t, actual.Msg == defaultSuccessResponse.Msg))
// 		assertionLogger.LogAssertion(t, "msgCode", "Update name failure - user not found", actual.MsgCode, defaultSuccessResponse.MsgCode,
// 			assert.True(t, actual.MsgCode == defaultSuccessResponse.MsgCode))
// 	})
// }
