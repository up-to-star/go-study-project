package repository

import (
	"basic_go/webook/internal/domain"
	"basic_go/webook/internal/repository/cache"
	"basic_go/webook/internal/repository/dao"
	"context"
	"database/sql"
	"fmt"
	"time"
)

var ErrUserDuplicateEmail = dao.ErrUserDuplicate
var ErrUserNotFound = dao.ErrUserNotFount

type UserRepository interface {
	FindById(ctx context.Context, id int64) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
	entityToDomain(u dao.User) domain.User
	domainToEntity(u domain.User) dao.User
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{dao: dao, cache: cache}
}

// FindById 屏蔽数据存储的逻辑，不管是cache还是mysql
func (r *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		return u, nil
	}

	u1, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = r.entityToDomain(u1)
	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			fmt.Println("redis 缓存设置失败")
		}
	}()

	return u, nil

	//return domain.User{}, cache.ErrKeyNotExist
}

func (r *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	ud := r.entityToDomain(u)
	return ud, nil
}

func (r *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	u, err := r.dao.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Nickname: u.Nickname,
		AboutMe:  u.AboutMe,
		WechatInfo: domain.WechatInfo{
			UnionId: u.WechatUnionId.String,
			OpenId:  u.WechatOpenId.String,
		},
		Birthday: time.UnixMilli(u.Birthday),
		Ctime:    time.UnixMilli(u.Ctime),
	}
}

func (r *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		WechatOpenId: sql.NullString{
			String: u.WechatInfo.OpenId,
			Valid:  u.WechatInfo.OpenId != "",
		},
		WechatUnionId: sql.NullString{
			String: u.WechatInfo.UnionId,
			Valid:  u.WechatInfo.UnionId != "",
		},
		Nickname: u.Nickname,
		AboutMe:  u.AboutMe,
		Password: u.Password,
		Ctime:    u.Ctime.UnixMilli(),
		Utime:    time.Now().UnixMilli(),
	}
}
