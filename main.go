package main

import (
	_ "autoTest/store/config"
	"autoTest/store/logger"
	"fmt"
)

func main() {
	// 初始化日志
	logger.InitLogger()
	// logger.Init(config.LogLevel)
	logger.Logger.Info("这是一个信息日志",
		"key", "value",
	)
	// 模拟一个错误
	err := someFunction()
	if err != nil {
		logger.LogError("报错消息", err)
	}
}

// someFunction 模拟一个返回错误的函数
func someFunction() error {
	return fmt.Errorf("出现错误")
}
