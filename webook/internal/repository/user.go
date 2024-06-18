package repository

import (
	"basic_go/webook/internal/domain"
	"basic_go/webook/internal/repository/dao"
	"context"
	"time"
)

var ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
var ErrUserNotFound = dao.ErrUserNotFount

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{dao: dao}
}

// FindById 屏蔽数据存储的逻辑，不管是cache还是mysql
func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		NickName: u.NickName,
		Password: u.Password,
		AboutMe:  u.AboutMe,
		Birthday: time.UnixMilli(u.Birthday),
		Ctime:    time.UnixMilli(u.Ctime),
	}, nil

}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{Email: u.Email, Password: u.Password})
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}
