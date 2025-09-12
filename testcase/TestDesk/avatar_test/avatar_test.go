package avatar_test

import (
	"autoTest/testcase"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	// Replace with your actual module path

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

var mergeLock sync.Mutex

// TestChangeAvatarSuccess 测试修改头像成功场景
func TestChangeAvatarSuccess(t *testing.T) {
	t.Parallel() // 启用并行测试
	runner.Run(t, "ChangeAvatarSuccess-"+t.Name(), func(p provider.T) {
		fmt.Println("Running TestChangeAvatarSuccess")
		p.Title("ChangeAvatarSuccess: 验证修改头像成功")
		p.Description("测试修改头像接口，确保 Code=200, Msg=success, MsgCode=0")
		// Try p.AddLabel if supported; otherwise, skip labels
		// p.AddLabel(allure.NewLabel("suite", "Avatar Tests"))
		// p.AddLabel(allure.NewLabel("environment", "staging"))

		// 执行登录作为前置处理器
		token, err := testcase.Login(t, p, "ChangeAvatarSuccess")
		if err != nil {
			t.Fatalf("登录失败: %v", err)
		}

		// 模拟修改头像请求
		mockRequest := func(token string) (testcase.Response, error) {
			fmt.Println("Executing ChangeAvatar request with token:", token)
			return testcase.Response{
				Code:    200,
				Msg:     "success",
				MsgCode: 0,
				Data:    map[string]string{"avatar_url": "https://example.com/avatar.jpg"},
			}, nil
		}

		// 调用辅助测试函数
		testcase.TestCommonResponse(t, p, mockRequest, "ChangeAvatarSuccess", allure.NORMAL, token)
	})
}

// TestChangeAvatarFailure 测试修改头像失败场景
func TestChangeAvatarFailure(t *testing.T) {
	t.Parallel() // 启用并行测试
	runner.Run(t, "ChangeAvatarFailure-"+t.Name(), func(p provider.T) {
		fmt.Println("Running TestChangeAvatarFailure")
		p.Title("ChangeAvatarFailure: 验证修改头像失败")
		p.Description("测试修改头像接口，期望 Code=200, Msg=success, MsgCode=0，但返回错误值")
		// Try p.AddLabel if supported; otherwise, skip labels
		// p.AddLabel(allure.NewLabel("suite", "Avatar Tests"))
		// p.AddLabel(allure.NewLabel("environment", "staging"))

		// 执行登录作为前置处理器
		token, err := testcase.Login(t, p, "ChangeAvatarFailure")
		if err != nil {
			t.Fatalf("登录失败: %v", err)
		}

		// 模拟修改头像失败请求
		mockRequest := func(token string) (testcase.Response, error) {
			fmt.Println("Executing ChangeAvatar failure request with token:", token)
			return testcase.Response{
				Code:    400,
				Msg:     "invalid avatar",
				MsgCode: 1,
			}, nil
		}

		// 调用辅助测试函数
		testcase.TestCommonResponse(t, p, mockRequest, "ChangeAvatarFailure", allure.CRITICAL, token)
	})
}

// TestMain 设置 Allure 环境变量并运行测试
func TestMain(m *testing.M) {
	// 设置模块内的临时 Allure 输出路径
	moduleResultsDir := "allure-results-avatar"
	os.Setenv("ALLURE_OUTPUT_FOLDER", moduleResultsDir)
	os.Setenv("ALLURE_OUTPUT_PATH", "./")
	fmt.Println("ALLURE_OUTPUT_FOLDER:", os.Getenv("ALLURE_OUTPUT_FOLDER"))
	os.MkdirAll(moduleResultsDir, 0755)

	// 运行测试
	code := m.Run()

	// 获取项目根目录（TestDesk）的绝对路径
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Println("无法获取工作目录:", err)
		os.Exit(1)
	}
	for filepath.Base(rootDir) != "TestDesk" {
		rootDir = filepath.Dir(rootDir)
		if rootDir == "/" || rootDir == "" {
			fmt.Println("无法找到 TestDesk 根目录")
			os.Exit(1)
		}
	}
	mergedDir := filepath.Join(rootDir, "allure-results")

	// 合并 Allure 结果到 TestDesk/allure-results
	mergeLock.Lock()
	defer mergeLock.Unlock()
	os.MkdirAll(mergedDir, 0755)
	files, err := filepath.Glob(filepath.Join(moduleResultsDir, "*.json"))
	if err != nil {
		fmt.Println("无法读取 Allure 结果文件:", err)
		os.Exit(1)
	}
	fmt.Println("合并文件:", files)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("无法读取文件:", file, err)
			continue
		}
		dest := filepath.Join(mergedDir, filepath.Base(file))
		if err := os.WriteFile(dest, data, 0644); err != nil {
			fmt.Println("无法写入合并文件:", dest, err)
		} else {
			fmt.Println("合并文件到:", dest)
		}
	}

	os.Exit(code)
}
