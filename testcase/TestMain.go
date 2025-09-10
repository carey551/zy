package testcase

import (
	"os"
	"testing"
)

// TestMain 设置 Allure 环境变量并运行测试
func TestMain(m *testing.M) {
	// 设置 Allure 结果输出路径
	os.Setenv("ALLURE_OUTPUT_FOLDER", "allure-results")
	os.Setenv("ALLURE_OUTPUT_PATH", "./")
	// 运行测试
	os.Exit(m.Run())
}
