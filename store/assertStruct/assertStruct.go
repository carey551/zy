package assertStruct

// 设置response响应的信息

type ResposeStruct struct {
	Code    int8   `json:"code"`
	Msg     string `json:"msg"`
	MsgCode int8   `json:"msgCode"`
}
