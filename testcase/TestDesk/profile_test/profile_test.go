package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"autoTest/testcase" // Replace with your actual module path

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

var mergeLock sync.Mutex

// TestUpdateProfileSuccess 测试更新用户详情成功场景
func TestUpdateProfileSuccess(t *testing.T) {
	t.Parallel() // 启用并行测试
	runner.Run(t, "UpdateProfileSuccess-"+t.Name(), func(p provider.T) {
		fmt.Println("Running TestUpdateProfileSuccess")
		p.Title("UpdateProfileSuccess: 验证更新用户详情成功")
		p.Description("测试更新用户详情接口，确保 Code=200, Msg=success, MsgCode=0")
		// Try p.AddLabel if supported; otherwise, skip labels
		// p.AddLabel(allure.NewLabel("suite", "Profile Tests"))
		// p.AddLabel(allure.NewLabel("environment", "staging"))

		// 执行登录作为前置处理器
		token, err := testcase.Login(t, p, "UpdateProfileSuccess")
		if err != nil {
			t.Fatalf("登录失败: %v", err)
		}

		// 模拟更新用户详情请求
		mockRequest := func(token string) (testcase.Response, error) {
			fmt.Println("Executing UpdateProfile request with token:", token)
			return testcase.Response{
				Code:    200,
				Msg:     "success",
				MsgCode: 0,
				Data:    map[string]string{"username": "new_user"},
			}, nil
		}

		// 调用辅助测试函数
		testcase.TestCommonResponse(t, p, mockRequest, "UpdateProfileSuccess", allure.NORMAL, token)
	})
}

// TestUpdateProfileFailure 测试更新用户详情失败场景
func TestUpdateProfileFailure(t *testing.T) {
	t.Parallel() // 启用并行测试
	runner.Run(t, "UpdateProfileFailure-"+t.Name(), func(p provider.T) {
		fmt.Println("Running TestUpdateProfileFailure")
		p.Title("UpdateProfileFailure: 验证更新用户详情失败")
		p.Description("测试更新用户详情接口，期望 Code=200, Msg=success, MsgCode=0，但返回错误值")
		// Try p.AddLabel if supported; otherwise, skip labels
		// p.AddLabel(allure.NewLabel("suite", "Profile Tests"))
		// p.AddLabel(allure.NewLabel("environment", "staging"))

		// 执行登录作为前置处理器
		token, err := testcase.Login(t, p, "UpdateProfileFailure")
		if err != nil {
			t.Fatalf("登录失败: %v", err)
		}

		// 模拟更新用户详情失败请求
		mockRequest := func(token string) (testcase.Response, error) {
			fmt.Println("Executing UpdateProfile failure request with token:", token)
			return testcase.Response{
				Code:    400,
				Msg:     "invalid profile data",
				MsgCode: 1,
			}, nil
		}

		// 调用辅助测试函数
		testcase.TestCommonResponse(t, p, mockRequest, "UpdateProfileFailure", allure.CRITICAL, token)
	})
}

// TestMain 设置 Allure 环境变量并运行测试
func TestMain(m *testing.M) {
	// 设置模块内的临时 Allure 输出路径
	moduleResultsDir := "allure-results-profile"
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
