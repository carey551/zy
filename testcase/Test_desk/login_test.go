package testdesk

import (
	"autoTest/deskSystem/login"
	"autoTest/store/assertStruct"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestA(t *testing.T) {
	// 正例：正确用户名和密码
	t.Run("SuccessfulLogin", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]string
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			assert.NoError(t, err)
			assert.Equal(t, "testuser", reqBody["username"])
			assert.Equal(t, "correctpassword", reqBody["password"])

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(assertStruct.ResposeStruct{
				Code:    0,
				Msg:     "success",
				MsgCode: 0,
			})
		}))
		defer server.Close()

		result, err := login.A(server.URL, "testuser", "correctpassword")
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, 200, result.Code, "unexpected code")
		assert.Equal(t, "success", result.Msg, "unexpected message")
		assert.Equal(t, "SUCCESS", result.MsgCode, "unexpected message code")
	})

	// 反例 1：错误凭据
	// t.Run("InvalidCredentials", func(t *testing.T) {
	// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		var reqBody map[string]string
	// 		err := json.NewDecoder(r.Body).Decode(&reqBody)
	// 		assert.NoError(t, err)
	// 		assert.Equal(t, "wronguser", reqBody["username"])
	// 		assert.Equal(t, "wrongpassword", reqBody["password"])

	// 		w.Header().Set("Content-Type", "application/json")
	// 		w.WriteHeader(http.StatusUnauthorized)
	// 		json.NewEncoder(w).Encode(LoginResponse{
	// 			Code:    401,
	// 			Msg:     "invalid credentials",
	// 			MsgCode: "ERR_AUTH",
	// 		})
	// 	}))
	// 	defer server.Close()

	// 	result, err := a(server.URL, "wronguser", "wrongpassword")
	// 	assert.NoError(t, err, "expected no error from client")
	// 	assert.Equal(t, 401, result.Code, "unexpected code")
	// 	assert.Equal(t, "invalid credentials", result.Msg, "unexpected message")
	// 	assert.Equal(t, "ERR_AUTH", result.MsgCode, "unexpected message code")
	//})

	// 反例 2：服务器 500 错误
	// t.Run("ServerError", func(t *testing.T) {
	// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	// 	}))
	// 	defer server.Close()

	// 	result, err := a(server.URL, "testuser", "correctpassword")
	// 	assert.Error(t, err, "expected an error due to server failure")
	// 	assert.Equal(t, 0, result.Code, "expected default code")
	// 	assert.Empty(t, result.Msg, "expected empty message")
	// 	assert.Empty(t, result.MsgCode, "expected empty message code")
	// })
}
