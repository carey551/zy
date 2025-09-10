package testdesk

import (
	cresponse "autoTest/store/commonResponse"
	"testing"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

// TestApiResponseSuccess 测试成功场景
func TestApiResponseSuccess(t *testing.T) {
	runner.Run(t, "TestSuccess", func(p provider.T) {
		p.Title("TestSuccess: Verify successful API response")
		p.Description("Tests the API response to ensure Code=200, Msg=success, MsgCode=0")
		// 模拟成功的请求函数
		mockRequest := func() (cresponse.CommonResponse, error) {
			return cresponse.CommonResponse{
				Code:    200,
				Msg:     "success",
				MsgCode: 0,
			}, nil
		}

		// 调用辅助测试函数，设置优先级为 NORMAL
		CtestApiResponse(t, p, mockRequest, "TestSuccess", allure.NORMAL)
	})
}

// TestApiResponseFailure 测试失败场景
func TestApiResponseFailure(t *testing.T) {
	runner.Run(t, "TestFailure", func(p provider.T) {
		p.Title("TestFailure: Verify API response with incorrect values")
		p.Description("Tests the API response expecting Code=200, Msg=success, MsgCode=0, but returns incorrect values")
		// 模拟失败的请求函数
		mockRequest := func() (cresponse.CommonResponse, error) {
			return cresponse.CommonResponse{
				Code:    400,
				Msg:     "error",
				MsgCode: 1,
			}, nil
		}

		// 调用辅助测试函数，设置优先级为 CRITICAL
		CtestApiResponse(t, p, mockRequest, "TestFailure", allure.CRITICAL)
	})
}
