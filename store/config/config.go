package config

import "time"

// 一些配置信息
const (
	AdminUrl      = ""              // 后台地址
	DeskURL       = ""              // 前台地址
	BetWingoUrl   = ""              // wingo地址
	LogLevel      = "INFO"          // 设置日志登记
	MAXWaitTime   = time.Second * 3 // 最大等待时间
	MAxRtryNumber = 3               // 最大重试次数
	FIXEDTime     = time.Second * 3 // 固定等待时间
)
