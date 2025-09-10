package testdesk

// 这个utils专门用来 处理testing和

import (
	cresponse "autoTest/store/commonResponse"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newLogger 初始化自定义格式的 zap 日志记录器
func newLogger() (*zap.Logger, error) {
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "name",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	return config.Build(zap.AddCaller())
}

// testApiResponse 辅助函数，用于执行实际的测试逻辑，带 zap 日志和 Allure 步骤
func CtestApiResponse(t *testing.T, p provider.T, requestFn func() (cresponse.CommonResponse, error), testName string, severity allure.SeverityType) {
	// 设置测试优先级
	p.Severity(severity)

	// 初始化 zap 日志
	logger, err := newLogger()
	if err != nil {
		t.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// 执行请求函数
	var resp cresponse.CommonResponse
	p.WithNewStep("Execute API request", func(sCtx provider.StepCtx) {
		sCtx.WithParameters(
			allure.NewParameter("Test Name", testName),
			allure.NewParameter("Expected Response", "Code=200, Msg=success, MsgCode=0"),
		)
		resp, err = requestFn()
		if err != nil {
			logger.Error("Request failed",
				zap.String("testName", testName),
				zap.Error(err))
			assert.NoError(t, err, "request failed")
			sCtx.FailNow()
			return
		}
		logger.Info("Request succeeded",
			zap.String("testName", testName),
			zap.Any("response", resp))
	})

	// 添加响应作为附件
	respBytes, _ := json.Marshal(resp)
	p.WithAttachments(allure.NewAttachment("API Response", allure.JSON, respBytes))

	// 期望的字段值
	const (
		expectedCode    = 200
		expectedMsg     = "success"
		expectedMsgCode = 0
	)

	// 收集断言失败的字段信息
	var failures []zap.Field
	pass := true

	// Allure 步骤：逐一断言每个字段
	p.WithNewStep("Assert Code field", func(sCtx provider.StepCtx) {
		sCtx.WithParameters(
			allure.NewParameter("Expected Code", fmt.Sprintf("%d", expectedCode)),
			allure.NewParameter("Actual Code", fmt.Sprintf("%d", resp.Code)),
		)
		if !assert.Equal(t, expectedCode, resp.Code, "Code field does not match") {
			pass = false
			failures = append(failures,
				zap.Int("expectedCode", expectedCode),
				zap.Int("actualCode", resp.Code))
			sCtx.FailNow()
		}
	})

	p.WithNewStep("Assert Msg field", func(sCtx provider.StepCtx) {
		sCtx.WithParameters(
			allure.NewParameter("Expected Msg", expectedMsg),
			allure.NewParameter("Actual Msg", resp.Msg),
		)
		if !assert.Equal(t, expectedMsg, resp.Msg, "Msg field does not match") {
			pass = false
			failures = append(failures,
				zap.String("expectedMsg", expectedMsg),
				zap.String("actualMsg", resp.Msg))
			sCtx.FailNow()
		}
	})

	p.WithNewStep("Assert MsgCode field", func(sCtx provider.StepCtx) {
		sCtx.WithParameters(
			allure.NewParameter("Expected MsgCode", fmt.Sprintf("%d", expectedMsgCode)),
			allure.NewParameter("Actual MsgCode", fmt.Sprintf("%d", resp.MsgCode)),
		)
		if !assert.Equal(t, expectedMsgCode, resp.MsgCode, "MsgCode field does not match") {
			pass = false
			failures = append(failures,
				zap.Int("expectedMsgCode", expectedMsgCode),
				zap.Int("actualMsgCode", resp.MsgCode))
			sCtx.FailNow()
		}
	})

	// 根据断言结果记录单条日志
	if pass {
		logger.Info("All assertions passed",
			zap.String("testName", testName),
			zap.Any("response", resp))
	} else {
		failureFields := append([]zap.Field{
			zap.String("testName", testName),
			zap.Any("response", resp),
		}, failures...)
		logger.Warn("Assertion failed", failureFields...)
	}
}
