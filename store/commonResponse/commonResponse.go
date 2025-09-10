package commonresponse

//公共的 设置response响应的信息
// Response 定义接口返回的响应结构体
type CommonResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	MsgCode int    `json:"msgCode"`
}
