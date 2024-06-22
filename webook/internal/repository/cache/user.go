package cache

import (
	"basic_go/webook/internal/domain"
	"context"
	"github.com/redis/go-redis/v9"
)

type UserCache struct {
	// 传单机Redis可以，传cluster 的redis也可以
	client redis.Cmdable
}

// A用到了B，B一定是接口
// A用到了B，B一定是A的字段
// A用到了B，A绝对不初始化B，而是外面注入

func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client: client,
	}
}

func (u *UserCache) GetUser(ctx context.Context, id int64) (domain.User, error {
	return domain.User{}, nil
}
