package Test_desk

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Response 定义返回结构体
type Response struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	MsgCode string `json:"msgCode"`
}

// 全局 Zap Logger
var logger *zap.Logger

// initLogger 初始化 Zap 日志记录器为开发模式
func initLogger() *zap.Logger {
	logger, err := zap.NewDevelopment() // 用于测试的人类可读日志
	if err != nil {
		panic(fmt.Sprintf("无法初始化 zap 日志记录器: %v", err))
	}
	return logger
}

// assertAndLog 包装 testify 断言以记录失败
func assertAndLog(t *testing.T, condition bool, msg string, fields ...zap.Field) bool {
	if !condition {
		logger.Error("断言失败",
			append([]zap.Field{
				zap.String("test", t.Name()),
				zap.String("message", msg),
			}, fields...)...)
	}
	return assert.New(t).True(condition, msg, fields)
}

// requireAndLog 包装 testify require 以记录失败
func requireAndLog(t *testing.T, condition bool, msg string, fields ...zap.Field) {
	if !condition {
		logger.Error("Require 断言失败",
			append([]zap.Field{
				zap.String("test", t.Name()),
				zap.String("message", msg),
			}, fields...)...)
	}
	require.New(t).True(condition, msg, fields)
}

// Login 模拟登录，返回 token
func Login(username, password string) (string, error) {
	logger.Info("尝试登录", zap.String("username", username))
	time.Sleep(100 * time.Millisecond) // 模拟延迟
	if username == "user" && password == "pass" {
		logger.Info("登录成功", zap.String("token", "valid_token_123"))
		return "valid_token_123", nil
	}
	logger.Error("登录失败", zap.String("username", username), zap.Error(fmt.Errorf("invalid credentials")))
	return "", fmt.Errorf("invalid credentials")
}

// GetUserDetails 模拟获取用户详情
func GetUserDetails(token string) (Response, error) {
	logger.Info("获取用户详情", zap.String("token", token))
	if token == "valid_token_123" {
		resp := Response{
			Code:    200,
			Msg:     "User details: name=John, age=30",
			MsgCode: "SUCCESS",
		}
		logger.Info("用户详情获取成功", zap.Any("response", resp))
		return resp, nil
	}
	logger.Error("获取用户详情失败", zap.String("token", token), zap.Error(fmt.Errorf("unauthorized: invalid token")))
	return Response{}, fmt.Errorf("unauthorized: invalid token")
}

// EnterForum 模拟进入论坛
func EnterForum(token string) (Response, error) {
	logger.Info("进入论坛", zap.String("token", token))
	if token == "valid_token_123" {
		resp := Response{
			Code:    200,
			Msg:     "Welcome to the forum!",
			MsgCode: "SUCCESS",
		}
		logger.Info("成功进入论坛", zap.Any("response", resp))
		return resp, nil
	}
	logger.Error("进入论坛失败", zap.String("token", token), zap.Error(fmt.Errorf("unauthorized: invalid token")))
	return Response{}, fmt.Errorf("unauthorized: invalid token")
}

// Logout 模拟 token 清理
func Logout(token string) error {
	logger.Info("正在注销", zap.String("token", token))
	return nil
}

// TestMain 用于全局前置/后置处理
func TestMain(m *testing.M) {
	// 初始化全局日志记录器
	logger = initLogger()
	defer logger.Sync()

	logger.Info("全局前置：初始化测试环境")
	code := m.Run()
	logger.Info("全局后置：清理测试环境")
	os.Exit(code)
}

// TestLoginAndGetUserDetails 测试登录和用户详情
func TestLoginAndGetUserDetails(t *testing.T) {
	t.Parallel()

	// 前置：登录
	logger.Info("设置 TestLoginAndGetUserDetails")
	token, err := Login("user", "pass")
	requireAndLog(t, err == nil, "前置失败：登录错误", zap.Error(err))

	// 后置：注销
	t.Cleanup(func() {
		logger.Info("清理 TestLoginAndGetUserDetails")
		err := Logout(token)
		assertAndLog(t, err == nil, "后置失败：注销错误", zap.Error(err))
	})

	// 测试：验证用户详情
	resp, err := GetUserDetails(token)
	assertAndLog(t, err == nil, "GetUserDetails 失败", zap.Error(err))
	assertAndLog(t, resp.Code == 200, "响应代码不符合预期", zap.Int("expected", 200), zap.Int("actual", resp.Code))
	// 模拟一个断言失败（仅用于演示）
	assertAndLog(t, resp.Msg == "Wrong details", "响应消息不符合预期",
		zap.String("expected", "Wrong details"), zap.String("actual", resp.Msg))
	assertAndLog(t, resp.MsgCode == "SUCCESS", "响应消息代码不符合预期",
		zap.String("expected", "SUCCESS"), zap.String("actual", resp.MsgCode))

	// 子测试：无效 token
	t.Run("SubTestInvalidToken", func(t *testing.T) {
		t.Parallel()
		logger.Info("运行 GetUserDetails 的 SubTestInvalidToken")
		resp, err := GetUserDetails("invalid_token")
		assertAndLog(t, err != nil, "应为无效 token 返回错误")
		assertAndLog(t, err.Error() == "unauthorized: invalid token", "错误消息不符合预期",
			zap.String("expected", "unauthorized: invalid token"), zap.String("actual", err.Error()))
		assertAndLog(t, resp == Response{}, "应为无效 token 返回空 Response",
			zap.Any("expected", Response{}), zap.Any("actual", resp))
	})
}

// TestLoginAndEnterForum 测试登录和论坛访问
func TestLoginAndEnterForum(t *testing.T) {
	t.Parallel()

	// 前置：登录
	logger.Info("设置 TestLoginAndEnterForum")
	token, err := Login("user", "pass")
	requireAndLog(t, err == nil, "前置失败：登录错误", zap.Error(err))

	// 后置：注销
	defer func() {
		logger.Info("清理 TestLoginAndEnterForum")
		err := Logout(token)
		assertAndLog(t, err == nil, "后置失败：注销错误", zap.Error(err))
	}()

	// 测试：验证论坛访问
	resp, err := EnterForum(token)
	assertAndLog(t, err == nil, "EnterForum 失败", zap.Error(err))
	assertAndLog(t, resp.Code == 200, "响应代码不符合预期", zap.Int("expected", 200), zap.Int("actual", resp.Code))
	assertAndLog(t, resp.Msg == "Welcome to the forum!", "响应消息不符合预期",
		zap.String("expected", "Welcome to the forum!"), zap.String("actual", resp.Msg))
	assertAndLog(t, resp.MsgCode == "SUCCESS", "响应消息代码不符合预期",
		zap.String("expected", "SUCCESS"), zap.String("actual", resp.MsgCode))

	// 子测试：无效 token
	t.Run("SubTestInvalidToken", func(t *testing.T) {
		t.Parallel()
		logger.Info("运行 EnterForum 的 SubTestInvalidToken")
		resp, err := EnterForum("invalid_token")
		assertAndLog(t, err != nil, "应为无效 token 返回错误")
		assertAndLog(t, err.Error() == "unauthorized: invalid token", "错误消息不符合预期",
			zap.String("expected", "unauthorized: invalid token"), zap.String("actual", err.Error()))
		assertAndLog(t, resp == Response{}, "应为无效 token 返回空 Response",
			zap.Any("expected", Response{}), zap.Any("actual", resp))
	})
}

// AuthFixture 用于可复用的前置/后置处理
// type AuthFixture struct {
// 	token string
// 	mu    sync.Mutex
// }

// // NewAuthFixture 创建 Fixture
// func NewAuthFixture(t *testing.T) *AuthFixture {
// 	t.Helper()
// 	logger.Info("设置 AuthFixture")
// 	token, err := Login("user", "pass")
// 	requireAndLog(t, err == nil, "Fixture 前置失败", zap.Error(err))
// 	f := &AuthFixture{token: token}
// 	t.Cleanup(func() {
// 		f.mu.Lock()
// 		defer f.mu.Unlock()
// 		logger.Info("清理 AuthFixture")
// 		err := Logout(f.token)
// 		assertAndLog(t, err == nil, "Fixture 后置失败", zap.Error(err))
// 	})
// 	return f
// }

// // TestLoginAndAccessBothWithFixture 测试用户详情和论坛
// func TestLoginAndAccessBothWithFixture(t *testing.T) {
// 	t.Parallel()

// 	// 前置：使用 Fixture
// 	f := NewAuthFixture(t)

// 	// 测试：用户详情
// 	logger.Info("在 TestLoginAndAccessBothWithFixture 中测试 GetUserDetails")
// 	respDetails, err := GetUserDetails(f.token)
// 	assertAndLog(t, err == nil, "GetUserDetails 失败", zap.Error(err))
// 	assertAndLog(t, respDetails == Response{
// 		Code:    200,
// 		Msg:     "User details: name=John, age=30",
// 		MsgCode: "SUCCESS",
// 	}, "用户详情响应不符合预期",
// 		zap.Any("expected", Response{Code: 200, Msg: "User details: name=John, age=30", MsgCode: "SUCCESS"}),
// 		zap.Any("actual", respDetails))

// 	// 测试：论坛
// 	logger.Info("在 TestLoginAndAccessBothWithFixture 中测试 EnterForum")
// 	respForum, err := EnterForum(f.token)
// 	assertAndLog(t, err == nil, "EnterForum 失败", zap.Error(err))
// 	assertAndLog(t, respForum == Response{
// 		Code:    200,
// 		Msg:     "Welcome to the forum!",
// 		MsgCode: "SUCCESS",
// 	}, "论坛响应不符合预期",
// 		zap.Any("expected", Response{Code: 200, Msg: "Welcome to the forum!", MsgCode: "SUCCESS"}),
// 		zap.Any("actual", respForum))
// }
