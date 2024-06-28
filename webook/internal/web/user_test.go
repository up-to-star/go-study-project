package web

import (
	"basic_go/webook/internal/domain"
	"basic_go/webook/internal/service"
	svcmocks "basic_go/webook/internal/service/mocks"
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEncrypt(t *testing.T) {
	password := "hello#world123"
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		t.Fatal(err)
	}

	err = bcrypt.CompareHashAndPassword(encrypted, []byte(password))
	assert.NoError(t, err)
}

func TestUserHandler_Signup(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wangBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(nil)
				return usersvc
			},
			reqBody: `
					{
						"email": "123@qq.com",
						"password": "hello#world123",
						"confirmPassword": "hello#world123"
					}`,
			wantCode: http.StatusOK,
			wangBody: "注册成功",
		},
		{
			name: "参数不对,bind失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				//usersvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(nil)
				return usersvc
			},
			reqBody: `
					{
						"email": "123@qq.com",
						"password": "hello#world123",
						"confirmPassword": "hello#world123"
					`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
					{
						"email": "123qq.com",
						"password": "hello#world123",
						"confirmPassword": "hello#world123"
					}`,
			wantCode: http.StatusOK,
			wangBody: "非法邮箱格式",
		},
		{
			name: "两次输入密码不一致",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
					{
						"email": "123@qq.com",
						"password": "hello#world1253",
						"confirmPassword": "hello#world123"
					}`,
			wantCode: http.StatusOK,
			wangBody: "两次密码不一致",
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
					{
						"email": "123@qq.com",
						"password": "123",
						"confirmPassword": "123"
					}`,
			wantCode: http.StatusOK,
			wangBody: "密码必须包含字母、数字、特殊字符, 并且不少于8位",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(service.ErrUserDuplicateEmail)
				return usersvc
			},
			reqBody: `
					{
						"email": "123@qq.com",
						"password": "hello#world123",
						"confirmPassword": "hello#world123"
					}`,
			wantCode: http.StatusOK,
			wangBody: "邮箱冲突",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 拿到响应
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			h := NewUserHandler(tc.mock(ctrl), nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			//h := NewUserHandler()
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wangBody, resp.Body.String())
		})
	}
}

func TestMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	usersvc := svcmocks.NewMockUserService(ctrl)
	usersvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(errors.New("mock error"))
	err := usersvc.Signup(context.Background(), domain.User{
		Email: "123@qq.com",
	})
	t.Log(err)
}
