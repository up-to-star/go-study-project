package cache

import (
	"basic_go/webook/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
	Key(id int64) string
}

type RedisCache struct {
	// 传单机Redis可以，传cluster 的redis也可以
	client     redis.Cmdable
	expiration time.Duration
}

// A用到了B，B一定是接口
// A用到了B，B一定是A的字段
// A用到了B，A绝对不初始化B，而是外面注入

func NewUserCache(client redis.Cmdable) UserCache {
	return &RedisCache{
		client:     client,
		expiration: time.Minute * 30,
	}
}

func (cache *RedisCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.Key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}

	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *RedisCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.Key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *RedisCache) Key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
