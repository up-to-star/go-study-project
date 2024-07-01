package ratelimit

import (
	"basic_go/webook/internal/service/sms"
	"basic_go/webook/pkg/limiter"
	"context"
	"errors"
)

var errLimited = errors.New("触发限流")

type RateLimitSMSService struct {
	svc     sms.Service
	limiter limiter.Limiter
	key     string
}

func NewRateLimitSMSService(svc sms.Service, l limiter.Limiter) *RateLimitSMSService {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: l,
		key:     "sms-limiter",
	}
}

func (r *RateLimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		return err
	}
	if limited {
		return errLimited
	}
	return r.svc.Send(ctx, tplId, args, numbers...)
}
