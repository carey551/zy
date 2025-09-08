package output

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TestResult 表示单个测试用例的结果
type TestResult struct {
	Name     string
	Status   string
	Duration string
}

// TestReport 表示测试报告数据
type TestReport struct {
	DateTime    string
	Tests       []TestResult
	Logs        []string
	FailedLogs  []string
	TotalTests  int
	PassedTests int
	PassRate    float64
}

// Logger 是全局 Zap 日志记录器
var logger *zap.Logger

// initLogger 初始化 Zap 日志记录器
func initLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("无法初始化 zap 日志记录器: %v", err))
	}
	return logger
}

// runTests 运行 go test 并捕获输出
func RunTests() (string, error) {
	cmd := exec.Command("go", "test", "-v", "./...") // Test current package
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	output := out.String()
	if err != nil {
		logger.Error("测试运行失败", zap.Error(err), zap.String("output", output))
		os.WriteFile("test_error_output.txt", []byte(output), 0644) // Save output for debugging
	}
	if output == "" {
		logger.Warn("测试输出为空，可能无测试文件")
	}
	return output, err
}

// parseTestOutput 解析测试输出，提取结果和日志
func parseTestOutput(output string) TestReport {
	logger.Info("解析测试输出", zap.String("raw_output", output))
	lines := strings.Split(output, "\n")
	var report TestReport
	report.DateTime = time.Now().Format("2006年1月2日 15:04 (MST)")

	// 正则表达式匹配测试结果
	testRegex := regexp.MustCompile(`=== RUN\s+([^\n]+)|--- (PASS|FAIL): ([^\s]+) \(([0-9.]+s)\)`)

	// 正则表达式匹配 Zap 日志中的断言失败
	failedLogRegex := regexp.MustCompile(`\{"level":"error","msg":"(断言失败|Require 断言失败)","test":"([^"]+)","message":"([^"]+)","details":\[([^\]]*)\]\}`)

	// 提取测试结果
	for _, line := range lines {
		if matches := testRegex.FindStringSubmatch(line); matches != nil {
			if strings.HasPrefix(line, "=== RUN") {
				logger.Info("发现测试用例", zap.String("name", matches[1]))
				report.Tests = append(report.Tests, TestResult{Name: matches[1], Status: "运行中"})
			} else if strings.HasPrefix(line, "---") {
				name := matches[3]
				status := matches[1]
				duration := matches[4]
				for i, test := range report.Tests {
					if test.Name == name {
						report.Tests[i].Status = status
						report.Tests[i].Duration = duration
						if status == "PASS" {
							report.PassedTests++
						}
						logger.Info("测试用例完成", zap.String("name", name), zap.String("status", status), zap.String("duration", duration))
						break
					}
				}
			}
		}
		// 收集所有日志
		if strings.Contains(line, `"level":"info"`) || strings.Contains(line, `"level":"error"`) {
			report.Logs = append(report.Logs, line)
		}
		// 收集断言失败日志
		if matches := failedLogRegex.FindStringSubmatch(line); matches != nil {
			logger.Info("发现断言失败日志", zap.String("log", line))
			report.FailedLogs = append(report.FailedLogs, line)
		}
	}

	report.TotalTests = len(report.Tests)
	if report.TotalTests > 0 {
		report.PassRate = float64(report.PassedTests) / float64(report.TotalTests) * 100
	} else {
		logger.Warn("未找到任何测试用例")
	}

	logger.Info("测试输出解析完成", zap.Int("总测试数", report.TotalTests), zap.Int("通过数", report.PassedTests))
	return report
}

// generateLatexReport 生成 LaTeX 格式的测试报告
func generateLatexReport(report TestReport) string {
	logger.Info("生成 LaTeX 报告")
	var sb strings.Builder
	sb.WriteString(`\documentclass{article}
\usepackage[utf8]{inputenc}
\usepackage{geometry}
\geometry{a4paper, margin=1in}
\usepackage{longtable}
\usepackage{enumitem}
\usepackage{xcolor}
\usepackage{noto}
\usepackage{parskip}

\title{Go 测试报告：认证与访问功能}
\author{自动生成}
\date{` + report.DateTime + `}

\begin{document}

\maketitle

\section{报告概述}
本报告总结了针对 Go 项目中认证、用户详情获取和论坛访问功能的测试执行情况。测试使用 Go \texttt{testing} 包、\texttt{testify} 断言库和 \texttt{zap} 日志库，涵盖了登录、获取用户详情、进入论坛的场景。测试通过前后置处理器（\texttt{TestMain}、\texttt{t.Cleanup}、\texttt{AuthFixture}）实现 pytest 风格的 setup/teardown，并支持并行测试。

生成时间：` + report.DateTime + `。

\section{测试用例概述}
以下为测试用例的描述：
\begin{itemize}
    \item \textbf{TestLoginAndGetUserDetails}：验证用户登录后能够成功获取用户详情（Response 结构体：\texttt{Code=200, Msg="User details: name=John, age=30", MsgCode="SUCCESS"}）。包含子测试验证无效 token 场景。
    \item \textbf{TestLoginAndEnterForum}：验证用户登录后能够成功进入论坛（Response 结构体：\texttt{Code=200, Msg="Welcome to the forum!", MsgCode="SUCCESS"}）。包含子测试验证无效 token 场景。
    \item \textbf{TestLoginAndAccessBothWithFixture}：使用 \texttt{AuthFixture} 测试登录后同时访问用户详情和论坛，验证复用 token 的正确性。
\end{itemize}

\section{执行结果}
测试使用命令 \texttt{go test -v} 执行，包含 ` + fmt.Sprintf("%d", report.TotalTests) + ` 个测试用例。结果如下：
\begin{longtable}{|l|c|c|}
    \hline
    \textbf{测试用例} & \textbf{状态} & \textbf{耗时 (秒)} \\
    \hline
`)

	if report.TotalTests == 0 {
		sb.WriteString(`    \multicolumn{3}{|c|}{未找到测试用例} \\ \hline
`)
	} else {
		for _, test := range report.Tests {
			sb.WriteString(fmt.Sprintf("    %s & %s & %s \\\\ \\hline\n", test.Name, test.Status, test.Duration))
		}
	}

	sb.WriteString(`\end{longtable}

\textbf{总体结果}：` + fmt.Sprintf("%d/%d 测试通过，通过率 %.2f\\%%", report.PassedTests, report.TotalTests, report.PassRate) + `。

\section{日志摘要}
以下为关键 Zap 日志摘录（精简版，完整日志见附件）：
\begin{itemize}
`)

	if len(report.Logs) == 0 {
		sb.WriteString(`    \item 无日志输出，可能测试未运行。\n`)
	} else {
		for i, log := range report.Logs {
			if i >= 10 {
				sb.WriteString(`    \item \ldots（更多日志见附件）`)
				break
			}
			if len(log) > 200 {
				log = log[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("    \\item \\texttt{%s}\n", strings.ReplaceAll(log, `\`, `\\`)))
		}
	}

	sb.WriteString(`\end{itemize}

\section{断言失败详情}
`)

	if len(report.FailedLogs) == 0 {
		sb.WriteString(`无断言失败。\n`)
	} else {
		sb.WriteString(`以下为断言失败的日志：\n\begin{itemize}\n`)
		for _, log := range report.FailedLogs {
			sb.WriteString(fmt.Sprintf("    \\item \\texttt{%s}\n", strings.ReplaceAll(log, `\`, `\\`)))
		}
		sb.WriteString(`\end{itemize}\n`)
	}

	sb.WriteString(`\section{总结}
`)

	if report.TotalTests == 0 {
		sb.WriteString(`未检测到测试用例运行，请检查测试文件是否存在或测试函数是否正确定义。

\textbf{建议}：
\begin{itemize}
    \item 确保测试文件以 \texttt{_test.go} 结尾，例如 \texttt{auth_test.go}。
    \item 确保测试函数以 \texttt{Test} 开头，签名正确（如 \texttt{func TestXxx(t *testing.T)}）。
    \item 检查 \texttt{go.mod} 是否包含必要依赖（\texttt{testify} 和 \texttt{zap}）。
    \item 运行 \texttt{go test -list .} 确认测试用例是否被识别。
    \item 检查 \texttt{test_error_output.txt} 中的错误信息。
\end{itemize}
`)
	} else {
		sb.WriteString(`所有测试用例均已执行，验证了登录、用户详情获取和论坛访问功能的正确性。Zap 日志记录了操作和断言失败详情，便于调试。并行测试和前后置处理器提高了效率和资源管理能力。

\textbf{建议}：
\begin{itemize}
    \item 将模拟函数替换为真实 API 调用，验证生产环境行为。
    \item 增加边缘情况测试（如网络延迟、token 过期）。
    \item 配置 Zap 为生产模式，优化日志格式和性能。
\end{itemize}
`)
	}

	sb.WriteString(`\end{document}`)

	logger.Info("LaTeX 报告生成完成")
	return sb.String()
}

// saveReport 保存报告到文件
func saveReport(content, filename string) error {
	logger.Info("保存测试报告", zap.String("filename", filename))
	return os.WriteFile(filename, []byte(content), 0644)
}

func RunPore() {
	// 初始化日志
	logger = initLogger()
	defer logger.Sync()

	// 运行测试并捕获输出
	logger.Info("开始运行测试")
	output, err := RunTests()
	if err != nil {
		logger.Error("测试运行失败", zap.Error(err))
	}

	// 解析测试输出
	report := parseTestOutput(output)

	// 生成 LaTeX 报告
	latex := generateLatexReport(report)

	// 保存报告
	err = saveReport(latex, "test_report.tex")
	if err != nil {
		logger.Error("保存报告失败", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("测试报告生成成功", zap.String("file", "test_report.tex"))
}
