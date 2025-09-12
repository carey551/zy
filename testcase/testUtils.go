package testcase

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Response 定义 API 响应结构体
type Response struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	MsgCode int         `json:"msgCode"`
	Data    interface{} `json:"data,omitempty"`
}

// NewLogger 初始化 zap 日志记录器，保存到 TestDesk/zapLogger/YYYYMMDD_HHMMSS.json
func NewLogger() (*zap.Logger, error) {
	// 获取 TestDesk 根目录
	rootDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("无法获取工作目录: %v", err)
	}
	for filepath.Base(rootDir) != "TestDesk" {
		rootDir = filepath.Dir(rootDir)
		if rootDir == "/" || rootDir == "" {
			return nil, fmt.Errorf("无法找到 TestDesk 根目录")
		}
	}

	// 配置日志目录和文件名
	logDir := filepath.Join(rootDir, "zapLogger")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建日志目录 %s: %v", logDir, err)
	}
	logFile := filepath.Join(logDir, time.Now().Format("20060102_150405")+".json")

	// 配置 lumberjack 用于日志分割
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100,   // 100MB
		MaxAge:     28,    // 28 天
		MaxBackups: 0,     // 保留所有备份
		Compress:   false, // 不压缩
	}

	// 自定义 caller 格式：模块名/函数名
	customCaller := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		// 获取调用栈信息
		pc, file, _, ok := runtime.Caller(5) // 调整 skip 层级以获取正确调用者
		if !ok {
			enc.AppendString("unknown")
			return
		}
		// 提取模块名（从文件路径）
		module := filepath.Base(filepath.Dir(file)) // 如 avatar, profile
		// 提取函数名
		funcName := runtime.FuncForPC(pc).Name()
		// 清理函数名，保留最后一部分
		parts := strings.Split(funcName, "/")
		if len(parts) > 0 {
			funcName = parts[len(parts)-1]
			// 去除包路径前缀（如 main.）
			if idx := strings.Index(funcName, "."); idx != -1 {
				funcName = funcName[idx+1:]
			}
		}
		enc.AppendString(fmt.Sprintf("%s/%s", module, funcName))
	}

	// 自定义时间格式：2006-01-02 15:04:05
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}

	// 配置 JSON 编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder, // 使用自定义时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   customCaller, // 使用自定义 caller
	}
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 创建 zap core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(lumberjackLogger),
		zapcore.DebugLevel,
	)

	// 创建 logger
	logger := zap.New(core, zap.AddCaller())
	//fmt.Println("初始化日志文件:", logFile)
	return logger, nil
}

// Login 模拟登录并返回 token
func Login(t *testing.T, p provider.T, testName string) (string, error) {
	logger, err := NewLogger()
	if err != nil {
		t.Fatalf("初始化 zap 日志记录器失败: %v", err)
	}
	defer logger.Sync()

	//fmt.Println("执行登录:", testName)
	//logger.Info("执行登录", zap.String("testName", testName))
	var token string
	p.WithNewStep("执行登录请求", func(sCtx provider.StepCtx) {
		sCtx.WithParameters(
			allure.NewParameter("测试名称", testName),
			allure.NewParameter("预期结果", "登录成功并返回 token"),
		)
		// 模拟登录请求
		resp := Response{
			Code:    200,
			Msg:     "success",
			MsgCode: 0,
			Data:    map[string]string{"token": "mock-token-12345"},
		}
		respBytes, _ := json.Marshal(resp)
		p.WithAttachments(allure.NewAttachment("登录响应", allure.JSON, respBytes))

		if resp.Code != 200 {
			logger.Error("登录失败",
				zap.String("testName", testName),
				zap.Any("response", resp))
			sCtx.FailNow()
			return
		}
		// logger.Info("登录成功",
		// 	zap.String("testName", testName),
		// 	zap.Any("response", resp))
		token = resp.Data.(map[string]string)["token"]
	})
	return token, nil
}

// TestCommonResponse 执行测试逻辑，每个用例记录一条日志（成功或失败）
func TestCommonResponse(t *testing.T, p provider.T, requestFn func(token string) (Response, error), testName string, severity allure.SeverityType, token string) {
	fmt.Println("开始 Allure 测试:", testName)
	// 设置测试优先级
	p.Severity(severity)

	// 初始化 zap 日志记录器
	logger, err := NewLogger()
	if err != nil {
		t.Fatalf("初始化 zap 日志记录器失败: %v", err)
	}
	defer logger.Sync()

	// 添加调试步骤
	p.WithNewStep("调试 Allure 初始化", func(sCtx provider.StepCtx) {
		//fmt.Println("Allure 步骤生成:", testName)
		sCtx.WithParameters(allure.NewParameter("调试", "检查 Allure 是否工作"))
	})

	// 执行请求函数
	var resp Response
	p.WithNewStep("执行 API 请求", func(sCtx provider.StepCtx) {
		sCtx.WithParameters(
			allure.NewParameter("测试名称", testName),
			allure.NewParameter("Token", token),
			allure.NewParameter("预期响应", "Code=200, Msg=success, MsgCode=0"),
		)
		resp, err = requestFn(token)
		if err != nil {
			logger.Error("请求失败",
				zap.String("testName", testName),
				zap.Error(err))
			assert.NoError(t, err, "请求失败")
			sCtx.FailNow()
			return
		}
	})

	// 将响应添加为附件
	respBytes, _ := json.Marshal(resp)
	p.WithAttachments(allure.NewAttachment("API 响应", allure.JSON, respBytes))

	// 预期的字段值
	const (
		expectedCode    = 200
		expectedMsg     = "success"
		expectedMsgCode = 0
	)

	// 收集断言失败的字段信息
	type failureDetail struct {
		Field    string      `json:"field"`
		Expected interface{} `json:"expected"`
		Actual   interface{} `json:"actual"`
	}
	var failures []failureDetail
	pass := true

	// Allure 步骤：断言 Code 字段
	p.WithNewStep("断言 Code 字段", func(sCtx provider.StepCtx) {
		sCtx.WithParameters(
			allure.NewParameter("预期 Code", fmt.Sprintf("%d", expectedCode)),
			allure.NewParameter("实际 Code", fmt.Sprintf("%d", resp.Code)),
		)
		if !assert.Equal(t, expectedCode, resp.Code, "Code 字段不匹配") {
			pass = false
			failures = append(failures, failureDetail{
				Field:    "Code",
				Expected: expectedCode,
				Actual:   resp.Code,
			})
		}
	})

	// Allure 步骤：断言 Msg 字段
	p.WithNewStep("断言 Msg 字段", func(sCtx provider.StepCtx) {
		sCtx.WithParameters(
			allure.NewParameter("预期 Msg", expectedMsg),
			allure.NewParameter("实际 Msg", resp.Msg),
		)
		if !assert.Equal(t, expectedMsg, resp.Msg, "Msg 字段不匹配") {
			pass = false
			failures = append(failures, failureDetail{
				Field:    "Msg",
				Expected: expectedMsg,
				Actual:   resp.Msg,
			})
		}
	})

	// Allure 步骤：断言 MsgCode 字段
	p.WithNewStep("断言 MsgCode 字段", func(sCtx provider.StepCtx) {
		sCtx.WithParameters(
			allure.NewParameter("预期 MsgCode", fmt.Sprintf("%d", expectedMsgCode)),
			allure.NewParameter("实际 MsgCode", fmt.Sprintf("%d", resp.MsgCode)),
		)
		if !assert.Equal(t, expectedMsgCode, resp.MsgCode, "MsgCode 字段不匹配") {
			pass = false
			failures = append(failures, failureDetail{
				Field:    "MsgCode",
				Expected: expectedMsgCode,
				Actual:   resp.MsgCode,
			})
		}
	})

	// 记录单条日志
	if pass {
		logger.Info("断言成功",
			zap.String("testName", testName),
			zap.Any("response", resp))
	} else {
		logger.Error("断言失败",
			zap.String("testName", testName),
			zap.Any("response", resp),
			zap.Any("failures", failures))
	}
}
