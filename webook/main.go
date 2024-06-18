package main

import (
	"basic_go/webook/internal/repository"
	"basic_go/webook/internal/repository/dao"
	"basic_go/webook/internal/service"
	"basic_go/webook/internal/web"
	"basic_go/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	server := initWebServer()
	//store := cookie.NewStore([]byte("secret"))
	//store := memstore.NewStore([]byte("uX6}oS1`eP0:jY0-oI9:oE4^wD2;tL4@"), []byte("zI1|eP7%tJ7_nD4%tK0;cB6.zU7~sT2>"))
	store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
		[]byte("uX6}oS1`eP0:jY0-oI9:oE4^wD2;tL4@"), []byte("zI1|eP7%tJ7_nD4%tK0;cB6.zU7~sT2>"))
	if err != nil {
		panic(err)
	}
	server.Use(sessions.Sessions("mysession", store))
	//server.Use(middleware.NewLoginMiddlewareBuilder().IgnorePaths("/users/login").
	//	IgnorePaths("/users/signup").Build())
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().IgnorePaths("/users/login").
		IgnorePaths("/users/signup").Build())
	db := initDB()
	u := initUser(db)
	u.RegisterRoutes(server)
	server.Run(":8080")
}

func initWebServer() *gin.Engine {
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
	return server
}

func initDB() *gorm.DB {
	dsn := "root:root@tcp(localhost:13316)/webook?charset=utf8mb4&parseTime=True&loc=Local"
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

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}
