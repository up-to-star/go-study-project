package web

import (
	"basic_go/webook/internal/domain"
	"basic_go/webook/internal/service"
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

const (
	emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	biz                  = "login"
)

// UserHandler 定义和用户有关的所有路由
type UserHandler struct {
	svc              service.UserService
	codeSvc          service.CodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:              svc,
		codeSvc:          codeSvc,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/signup", u.Signup)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	ug.POST("/login_sms/code/send", u.SendLoginSmsCode)
	ug.POST("/login_sms", u.LoginSMS)
}

func (u *UserHandler) Signup(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	// Bind 方法根据content-type解析数据到req中
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 数据库操作
	//ctx.String(http.StatusOK, "注册成功")
	//fmt.Println(req)
	isEmail, err := u.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不一致")
		return
	}
	isPassword, err := u.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符, 并且不少于8位")
		return
	}
	err = u.svc.Signup(ctx, domain.User{Email: req.Email, Password: req.Password})
	if errors.Is(err, service.ErrUserDuplicateEmail) {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	// 设置session
	sess := sessions.Default(ctx)
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		//Secure: true,
		HttpOnly: true,
		MaxAge:   60,
	})
	sess.Save()
	//ctx.String(http.StatusOK, "登录成功")
	ctx.JSON(http.StatusOK, &Result{
		Code:    0,
		Message: "登录成功",
	})
	return
}

func (u *UserHandler) Edit(ctx *gin.Context) {

}

func (u *UserHandler) Profile(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	id := sess.Get("userId").(int64)

	user, err := u.svc.Profile(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code:    1,
			Message: "用户信息不存在",
		})
		return
	}
	type Response struct {
		Id       int64
		Email    string
		Nickname string
		Birthday string
		AboutMe  string
	}
	ctx.JSON(http.StatusOK, &Result{
		Code:    0,
		Message: "ok",
		Data: &Response{
			Id:       user.Id,
			Email:    user.Email,
			Nickname: user.Nickname,
			Birthday: user.Birthday.Format("2006-01-02 15:04:05"),
		},
	})
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	sess.Save()
	ctx.JSON(http.StatusOK, &Result{
		Code:    0,
		Message: "退出登录成功",
	})
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	switch {
	case err == nil:
		u.setJWTToken(ctx, user.Id)
		ctx.JSON(http.StatusOK, &Result{
			Code:    0,
			Message: "登录成功",
		})
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		ctx.JSON(http.StatusOK, &Result{
			Code:    1,
			Message: "用户名或密码不对",
		})
	default:
		ctx.JSON(http.StatusOK, &Result{
			Code:    2,
			Message: "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, &Result{
		Code:    0,
		Message: "登录成功",
	})
	return
}

func (u *UserHandler) setJWTToken(ctx *gin.Context, uid int64) {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("uX6}oS1`eP0:jY0-oI9:oE4^wD2;tL4@"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
	}
	ctx.Header("x-jwt-token", tokenStr)
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, ok := ctx.Get("claims")
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}
	id := claims.Uid

	user, err := u.svc.Profile(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code:    1,
			Message: "用户信息不存在",
		})
		return
	}
	type Response struct {
		Id       int64
		Email    string
		Nickname string
		Birthday string
		AboutMe  string
	}
	//fmt.Println(user.Nickname)
	ctx.JSON(http.StatusOK, &Result{
		Code:    0,
		Message: "ok",
		Data: &Response{
			Id:       user.Id,
			Email:    user.Email,
			Nickname: user.Nickname,
			Birthday: user.Birthday.Format("2006-01-02 15:04:05"),
		},
	})
}

func (u *UserHandler) SendLoginSmsCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, &Result{
			Code:    1,
			Message: "手机号码为空",
		})
	}
	err := u.codeSvc.Send(ctx, biz, req.Phone)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, &Result{
			Code:    0,
			Message: "发送验证码成功",
		})
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, &Result{
			Code:    2,
			Message: "短信发送太频繁, 请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, &Result{
			Code:    3,
			Message: "系统错误",
		})
	}
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code:    3,
			Message: "系统异常",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, &Result{
			Code:    4,
			Message: "验证码不对，请重新输入",
		})
		return
	}
	ud, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code:    3,
			Message: "系统化错误",
		})
		return
	}
	u.setJWTToken(ctx, ud.Id)
	ctx.JSON(http.StatusOK, Result{
		Code:    0,
		Message: "登录成功",
	})
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
