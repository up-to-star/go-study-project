package main

import (
	"basic_go/webook/config"
	"basic_go/webook/internal/repository"
	"basic_go/webook/internal/repository/cache"
	"basic_go/webook/internal/repository/dao"
	"basic_go/webook/internal/service"
	"basic_go/webook/internal/service/oauth2/wechat"
	"basic_go/webook/internal/service/sms/localsms"
	"basic_go/webook/internal/service/sms/ratelimit"
	"basic_go/webook/internal/web"
	ijwt "basic_go/webook/internal/web/jwt"
	"basic_go/webook/internal/web/middleware"
	"basic_go/webook/pkg/limiter"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

func main() {
	db := initDB()
	rdb := initRedis()
	jhd := ijwt.NewRedisJWTHandler(rdb)
	server := initWebServer(jhd)
	u := initUser(db, rdb, jhd)
	u.RegisterRoutes(server)
	whd := initWechat("appid", "appSecrect", db, rdb, jhd)
	whd.RegisterRoutes(server)
	//server := gin.Default()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "你好，你来了")
	})
	server.Run(":8080")
}

func initWebServer(jhd ijwt.Handler) *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"PUT", "PATCH", "POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 让前端拿到token
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			//return origin == "https://github.com"
			if strings.Contains(origin, "http://localHost") {
				return true
			}
			return strings.Contains(origin, "xxx.com")
		},
		MaxAge: 12 * time.Hour,
	}))
	//store := cookie.NewStore([]byte("secret"))
	store := memstore.NewStore([]byte("uX6}oS1`eP0:jY0-oI9:oE4^wD2;tL4@"), []byte("zI1|eP7%tJ7_nD4%tK0;cB6.zU7~sT2>"))
	//store, err := redis.NewStore(16, "tcp", config.Config.Redis.Addr, "",
	//	[]byte("uX6}oS1`eP0:jY0-oI9:oE4^wD2;tL4@"), []byte("zI1|eP7%tJ7_nD4%tK0;cB6.zU7~sT2>"))
	//if err != nil {
	//	panic(err)
	//}
	server.Use(sessions.Sessions("mysession", store))
	//server.Use(middleware.NewLoginMiddlewareBuilder().IgnorePaths("/users/login").
	//	IgnorePaths("/users/signup").Build())
	server.Use(middleware.NewLoginJWTMiddlewareBuilder(jhd).IgnorePaths("/users/login").
		IgnorePaths("/users/signup").IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms").
		IgnorePaths("/oauth2/wechat/authurl").
		IgnorePaths("oauth2/wechat/callback").
		Build())
	return server
}

func initDB() *gorm.DB {
	dsn := config.Config.DB.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	return rdb
}

func initUser(db *gorm.DB, rdb *redis.Client, jhd ijwt.Handler) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	rd := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(ud, rd)
	svc := service.NewUserService(repo)
	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	//smsClient, err := sms.NewClient(common.NewCredential("", ""), "ap-nanjing", profile.NewClientProfile())
	//if err != nil {
	//	panic("smsClient初始化失败")
	//}
	//smsSvc := tencent.NewService(smsClient, "", "")
	smsSvc := localsms.NewService()
	rateSvc := ratelimit.NewRateLimitSMSService(smsSvc, limiter.NewRedisSlidingWindowLimiter(rdb, time.Second, 10))
	codeSvc := service.NewCodeService(codeRepo, rateSvc)
	u := web.NewUserHandler(svc, codeSvc, jhd)
	return u
}

func initWechat(appId string, appSecrect string, db *gorm.DB, rdb *redis.Client, jhd ijwt.Handler) *web.OAuth2WechatHandler {
	svc := wechat.NewService(appId, appSecrect)
	ud := dao.NewUserDAO(db)
	rd := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(ud, rd)
	userSvc := service.NewUserService(repo)
	return web.NewOAuth2WechatHandler(svc, userSvc, jhd)
}
