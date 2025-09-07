package login

import (
	"autoTest/store/assertStruct"
	"autoTest/store/logger"
)

func A(url, username, password string) (assertStruct.ResposeStruct, error) {
	var err error
	logger.Logger.Info(url, username, password)
	return assertStruct.ResposeStruct{
		Code:    0,
		Msg:     "",
		MsgCode: 0,
	}, err
}
