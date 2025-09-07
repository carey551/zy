package request

// 各种请求的封装
import (
	"autoTest/store/config"
	"autoTest/store/logger"
	"autoTest/store/utils"
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"golang.org/x/net/http2"
)

// get请求
/**
url 请求地址
api 接口
args[0] 添加额外的请求头设置
args[1] 添加参数
*/
func GetRequest(base_url, api string, args ...map[string]interface{}) ([]byte, *http.Response, error) {
	urlapi := base_url + api
	// 携带了参数需要设置参数
	if len(args[1]) > 0 {
		params := url.Values{}
		par := setGetParms(&params, args[1])
		urlapi = urlapi + "?" + par.Encode()
	}
	// 创建 GET 请求
	req, err := retryablehttp.NewRequest("GET", urlapi, nil)
	if err != nil {
		logger.LogError("创建Get请求失败", err)
		return nil, nil, err
	}
	// 设置请求头
	if len(args) > 0 {
		setHeaders(req, args[0])
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "keep-alive")

	// 获取请求句柄
	client := NewRetryableHTTPClient(config.MAxRtryNumber, config.FIXEDTime, config.MAXWaitTime)
	resp, err := client.Do(req)
	if err != nil {
		logger.LogError("发送Get请求失败", err)
		return nil, nil, err
	}
	defer resp.Body.Close()
	return handlerCode(resp)
}

/*
*
paylaod 请求参数 map[string]interface{}
base_url 请求地址
api 接口地址
args[0] 添加自定义请求头 map[string]interface{}
*/
func PostRequestCofig(payload map[string]interface{}, base_url, api string, args ...map[string]interface{}) ([]byte, *http.Response, error) {
	url := base_url + api
	// fmt.Printf("本次请求的地址%v\n", url)
	// 判断传进来的paylaod是否有签名，没有就添加上
	_, exists := payload["signature"]
	if !exists {
		payload["signature"] = ""
	}
	verfiyp := ""
	signature := utils.GetSignature(payload, &verfiyp)
	if signature == "" {
		payload["signature"] = signature
	}

	// fmt.Printf("请求的payload%v\n", payload)
	//将请求数据转换成json
	body, err := json.Marshal(payload)
	if err != nil {
		logger.LogError("json 编码失败", err)
		return nil, nil, err
	}
	// 发送post请求
	req, err := retryablehttp.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		logger.LogError("post请求失败", err)
		return nil, nil, err
	}
	// 设置请求头
	if len(args) > 0 {
		setHeaders(req, args[0])
	}
	req.Header.Set(
		"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
	)
	req.Header.Set("Content-Type", "application/problem+json; charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Connection", "keep-alive")
	client := NewRetryableHTTPClient(config.MAxRtryNumber, config.FIXEDTime, config.MAXWaitTime)

	// 打印所有请求头
	// for key, values := range req.Header {
	// 	for _, value := range values {
	// 		fmt.Printf("%s: %s\n", key, value)
	// 	}
	// }// fmt.Println("请求头:")

	//发送请求
	resp, err := client.Do(req)
	if err != nil {
		logger.LogError("发送post请求失败", err)
		return nil, nil, err
	}

	defer resp.Body.Close()
	// fmt.Println("响应状态码", resp.StatusCode)
	respBody, resp, err := handlerCode(resp)
	if err != nil {
		logger.LogError("post响应码处理失败", err)
		return nil, nil, err
	}

	return respBody, resp, nil
}

// 设置请求头
func setHeaders(req *retryablehttp.Request, headers map[string]interface{}) {
	for key, value := range headers {
		// 将 interface{} 转换为 string
		var headerValue string
		switch v := value.(type) {
		case string:
			headerValue = v
		case fmt.Stringer:
			headerValue = v.String()
		default:
			headerValue = fmt.Sprintf("%v", v)
		}
		req.Header.Set(key, headerValue)
	}
}

// 设置get参数
func setGetParms(params *url.Values, paramsMap map[string]interface{}) url.Values {
	for key, value := range paramsMap {
		var paramsValue string
		switch v := value.(type) {
		case string:
			paramsValue = v
		case fmt.Stringer:
			paramsValue = v.String()
		default:
			paramsValue = fmt.Sprintf("%v", v)
		}
		params.Add(key, paramsValue)
	}
	return *params
}

// 响应码的处理
func handlerCode(resp *http.Response) ([]byte, *http.Response, error) {
	if resp.StatusCode == 200 {
		//获取相应的内容
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.LogError("读取响应失败：", err)
			return nil, nil, err
		}
		return respBody, resp, nil

	} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		code := "状态码是" + string(resp.Status)
		errString := errors.New(code)
		logger.Logger.Warn("状态码是", resp.StatusCode)
		return nil, resp, errString
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		code := "状态码是" + string(resp.Status)
		errString := errors.New(code)
		logger.Logger.Warn("状态码是", resp.StatusCode)
		return nil, resp, errString
	} else {
		err := errors.New("状态码不是200~~或者是服务器错误~~~")
		logger.LogError("状态码：", err)
		return nil, nil, err
	}
}

// 验证请求的协议是不是http/2
func checkHttp2() *http.Client {
	// client := &http.Client{
	// 	// 检查 确保使用http/2
	// 	Transport: &http.Transport{
	// 		ForceAttemptHTTP2: true,
	// 	},
	// }
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 仅限测试，跳过证书验证
			},
		},
	}
	return client
}

// 三方库封装的要求支持最大重连，最大超时
func NewRetryableHTTPClient(
	retryMax int, // 最大重试次数
	wait time.Duration, // 固定重试等待时间
	timeout time.Duration, // 请求超时时间
) *retryablehttp.Client {
	// 自定义 Transport，开启 HTTP/2
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12, // TLS1.2 起
		},
	}
	// 启用 HTTP/2
	_ = http2.ConfigureTransport(transport)

	// 创建 retryablehttp.Client
	client := retryablehttp.NewClient()
	client.RetryMax = retryMax
	client.RetryWaitMin = wait
	client.RetryWaitMax = wait // 固定等待时间
	client.HTTPClient = &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return client
}

// 获取微妙级的时间戳并返回
func GetNowTime() int64 {
	now := time.Now()
	timestampMicro := now.Unix()
	return timestampMicro
}

func RandmoNine() int64 {
	//生成9位的随机数
	max_number := big.NewInt(900000000000)
	n, err := rand.Int(rand.Reader, max_number)
	if err != nil {
		logger.LogError("生成随机数失败：%v", err)
		return -1
	}
	random_number := n.Int64() + 100000000000
	return random_number
}
