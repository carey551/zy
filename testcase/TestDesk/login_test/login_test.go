package testdesk

// import (
// 	"fmt"
// 	"os"
// 	"testing"

// 	"autoTest/testcase"

// 	"github.com/ozontech/allure-go/pkg/allure"
// 	"github.com/ozontech/allure-go/pkg/framework/provider"
// 	"github.com/ozontech/allure-go/pkg/framework/runner"
// )

// // TestApiResponseSuccess 测试成功场景
// func TestApiResponseSuccess(t *testing.T) {
// 	runner.Run(t, "TestSuccess-"+t.Name(), func(p provider.T) {
// 		fmt.Println("Running TestApiResponseSuccess")
// 		p.Title("TestSuccess: 验证成功的 API 响应")
// 		p.Description("测试 API 响应，确保 Code=200, Msg=success, MsgCode=0")
// 		// 模拟成功的请求函数
// 		mockRequest := func() (testcase.Response, error) {
// 			fmt.Println("Executing mockRequest for TestSuccess")
// 			return testcase.Response{
// 				Code:    200,
// 				Msg:     "success",
// 				MsgCode: 0,
// 			}, nil
// 		}

// 		// 调用辅助测试函数，设置优先级为 NORMAL
// 		testcase.TestCommonResponse(t, p, mockRequest, "TestSuccess", allure.NORMAL)
// 	})
// }

// // TestApiResponseFailure 测试失败场景
// func TestApiResponseFailure(t *testing.T) {
// 	runner.Run(t, "TestFailure-"+t.Name(), func(p provider.T) {
// 		fmt.Println("Running TestApiResponseFailure")
// 		p.Title("TestFailure: 验证带有错误值的 API 响应")
// 		p.Description("测试 API 响应，期望 Code=200, Msg=success, MsgCode=0，但返回错误值")
// 		// 模拟失败的请求函数
// 		mockRequest := func() (testcase.Response, error) {
// 			fmt.Println("Executing mockRequest for TestFailure")
// 			return testcase.Response{
// 				Code:    400,
// 				Msg:     "error",
// 				MsgCode: 1,
// 			}, nil
// 		}

// 		// 调用辅助测试函数，设置优先级为 CRITICAL
// 		testcase.TestCommonResponse(t, p, mockRequest, "TestFailure", allure.CRITICAL)
// 	})
// }

// // TestMain 设置 Allure 环境变量并运行测试
// func TestMain(m *testing.M) {
// 	// 设置 Allure 输出路径
// 	os.Setenv("ALLURE_OUTPUT_FOLDER", "allure-results")
// 	os.Setenv("ALLURE_OUTPUT_PATH", "./")
// 	fmt.Println("ALLURE_OUTPUT_FOLDER:", os.Getenv("ALLURE_OUTPUT_FOLDER"))
// 	fmt.Println("ALLURE_OUTPUT_PATH:", os.Getenv("ALLURE_OUTPUT_PATH"))
// 	// 运行测试
// 	os.Exit(m.Run())
// }
