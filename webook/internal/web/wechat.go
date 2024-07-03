package web

import (
	"basic_go/webook/internal/service"
	"basic_go/webook/internal/service/oauth2/wechat"
	ijwt "basic_go/webook/internal/web/jwt"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	ijwt.Handler
	stateKey []byte
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService, hdl ijwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:      svc,
		userSvc:  userSvc,
		stateKey: []byte("uX6}oS1`eP0:jY0-oI9:oE4^wD2;tLs@"),
		Handler:  hdl,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	state := uuid.New().String()
	url, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code:    1,
			Message: "构造失败",
		})
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 3)),
		},
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code:    1,
			Message: "系统错误",
		})
	}
	ctx.SetCookie("jwt-state", tokenStr, 600, "/oauth2/wechat/callback",
		"", false, true)

	ctx.JSON(http.StatusOK, &Result{
		Code:    0,
		Message: "success",
		Data:    url,
	})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code:    5,
			Message: "登录失败",
		})
	}

	info, err := h.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code:    1,
			Message: "系统错误",
		})
		return
	}
	u, err := h.userSvc.FindOrCreateByWechat(ctx, info)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code:    1,
			Message: "系统错误",
		})
		return
	}
	err = h.SetLoginToken(ctx, u.Id)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, &Result{
		Message: "OK",
	})
}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie("jwt-state")
	if err != nil {
		return fmt.Errorf("拿不到 state cookie %s", err)
	}

	var sc StateClaims
	tokenStr, err := jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !tokenStr.Valid {
		return fmt.Errorf("token-state过期, %s", err)
	}
	if sc.State != state {
		return errors.New("state 不同")
	}
	return nil
}

type StateClaims struct {
	State string
	jwt.RegisteredClaims
}
