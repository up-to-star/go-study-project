package repository

import (
	"basic_go/webook/internal/domain"
	"basic_go/webook/internal/repository/cache"
	"basic_go/webook/internal/repository/dao"
	"context"
	"fmt"
	"time"
)

var ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
var ErrUserNotFound = dao.ErrUserNotFount

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{dao: dao, cache: cache}
}

// FindById 屏蔽数据存储的逻辑，不管是cache还是mysql
func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		return u, nil
	}

	u1, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = domain.User{
		Id:       u1.Id,
		Email:    u1.Email,
		Nickname: u1.Nickname,
		Password: u1.Password,
		AboutMe:  u1.AboutMe,
		Birthday: time.UnixMilli(u1.Birthday),
		Ctime:    time.UnixMilli(u1.Ctime),
	}
	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			fmt.Println("redis 缓存设置失败")
		}
	}()

	return u, nil

	//return domain.User{}, cache.ErrKeyNotExist
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{Email: u.Email, Password: u.Password})
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	ud := domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		Nickname: u.Nickname,
		AboutMe:  u.AboutMe,
		Birthday: time.UnixMilli(u.Birthday),
		Ctime:    time.UnixMilli(u.Ctime),
	}
	return ud, nil
}
