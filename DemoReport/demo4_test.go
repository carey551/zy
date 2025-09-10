package demoreport

// // 通用的带有日志信息的
// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/zap"
// 	"go.uber.org/zap/zapcore"
// )

// // Response 定义接口返回的响应结构体
// type Response struct {
// 	Code    int    `json:"code"`
// 	Msg     string `json:"msg"`
// 	MsgCode int    `json:"msgCode"`
// }

// // newLogger 初始化自定义格式的 zap 日志记录器
// func newLogger() (*zap.Logger, error) {
// 	config := zap.Config{
// 		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel), // 设置日志级别为 Debug（记录所有级别）
// 		Development: false,                                    // 生产模式，非开发模式
// 		Encoding:    "json",                                   // JSON 格式输出
// 		EncoderConfig: zapcore.EncoderConfig{
// 			TimeKey:        "timestamp",
// 			LevelKey:       "level",
// 			NameKey:        "name",
// 			CallerKey:      "caller",
// 			MessageKey:     "msg",
// 			StacktraceKey:  "stacktrace",
// 			LineEnding:     zapcore.DefaultLineEnding,
// 			EncodeLevel:    zapcore.LowercaseLevelEncoder, // level 字段小写（如 "info"）
// 			EncodeTime:     zapcore.ISO8601TimeEncoder,    // 时间格式如 "2025-09-10T21:27:40.036+0530"
// 			EncodeDuration: zapcore.SecondsDurationEncoder,
// 			EncodeCaller:   zapcore.ShortCallerEncoder, // caller 格式如 "test/api_test.go:53"
// 		},
// 		OutputPaths:      []string{"stdout"}, // 输出到控制台
// 		ErrorOutputPaths: []string{"stderr"}, // 错误输出到 stderr
// 	}
// 	return config.Build(zap.AddCaller()) // 启用 caller 信息
// }

// // testApiResponse 辅助函数，用于执行实际的测试逻辑，带 zap 日志
// func testApiResponse(t *testing.T, requestFn func() (Response, error), testName string) {
// 	// 初始化 zap 日志
// 	logger, err := newLogger()
// 	if err != nil {
// 		t.Fatalf("failed to initialize zap logger: %v", err)
// 	}
// 	defer logger.Sync() // 确保日志在测试结束时刷新

// 	t.Run(testName, func(t *testing.T) {
// 		// 执行请求函数
// 		resp, err := requestFn()
// 		if err != nil {
// 			logger.Error("Request failed",
// 				zap.String("testName", testName),
// 				zap.Error(err))
// 			assert.NoError(t, err, "request failed")
// 			return
// 		}

// 		// 期望的字段值
// 		const (
// 			expectedCode    = 200
// 			expectedMsg     = "success"
// 			expectedMsgCode = 0
// 		)

// 		// 收集断言失败的字段信息
// 		var failures []zap.Field
// 		pass := true

// 		// 逐一断言每个字段
// 		if !assert.Equal(t, expectedCode, resp.Code, "Code field does not match") {
// 			pass = false
// 			failures = append(failures,
// 				zap.Int("expectedCode", expectedCode),
// 				zap.Int("actualCode", resp.Code))
// 		}
// 		if !assert.Equal(t, expectedMsg, resp.Msg, "Msg field does not match") {
// 			pass = false
// 			failures = append(failures,
// 				zap.String("expectedMsg", expectedMsg),
// 				zap.String("actualMsg", resp.Msg))
// 		}
// 		if !assert.Equal(t, expectedMsgCode, resp.MsgCode, "MsgCode field does not match") {
// 			pass = false
// 			failures = append(failures,
// 				zap.Int("expectedMsgCode", expectedMsgCode),
// 				zap.Int("actualMsgCode", resp.MsgCode))
// 		}

// 		// 根据断言结果记录单条日志
// 		if pass {
// 			logger.Info("All assertions passed",
// 				zap.String("testName", testName),
// 				zap.Any("response", resp))
// 		} else {
// 			failureFields := append([]zap.Field{
// 				zap.String("testName", testName),
// 				zap.Any("response", resp),
// 			}, failures...)
// 			logger.Warn("Assertion failed", failureFields...)
// 		}
// 	})
// }

// // TestApiResponseSuccess 测试成功场景
// func TestApiResponseSuccess(t *testing.T) {
// 	// 模拟成功的请求函数
// 	mockRequest := func() (Response, error) {
// 		return Response{
// 			Code:    200,
// 			Msg:     "success",
// 			MsgCode: 0,
// 		}, nil
// 	}

// 	// 调用辅助测试函数
// 	testApiResponse(t, mockRequest, "TestSuccess")
// }

// // TestApiResponseFailure 测试失败场景
// func TestApiResponseFailure(t *testing.T) {
// 	// 模拟失败的请求函数
// 	mockRequest := func() (Response, error) {
// 		return Response{
// 			Code:    400,
// 			Msg:     "error",
// 			MsgCode: 1,
// 		}, nil
// 	}

// 	// 调用辅助测试函数
// 	testApiResponse(t, mockRequest, "TestFailure")
// }
