package redisembedded

import (
	"context"
	"errors"

	"github.com/bayu-gara/parking-lot/pkg/config"
	"github.com/bayu-gara/parking-lot/pkg/redis/embedded/sugardb"
)

type RedisEmbedded interface {
	Get(ctx context.Context, key string) (string, error)
	GetObj(ctx context.Context, key string, dest interface{}) error
	MGet(ctx context.Context, keys ...string) (result []string, err error)
	SetEX(ctx context.Context, key string, value string, expire int) error
	SetEXObj(ctx context.Context, key string, value interface{}, expire int) error
	Delete(ctx context.Context, keys ...string) (int64, error)
	SetNX(ctx context.Context, key string, value string, expire int) (bool, error)
	LPush(ctx context.Context, key string, values ...string) (int64, error)
	LPop(ctx context.Context, key string, count uint) ([]string, error)
	RPush(ctx context.Context, key string, values ...string) (int64, error)
	RPop(ctx context.Context, key string, count uint) ([]string, error)
}

func Init(cfg config.RedisEmbeddedConfig) (RedisEmbedded, error) {
	switch cfg.Engine {
	case "sugardb":
		return sugardb.Init()
	}

	return nil, errors.New("Unsupported database engine: " + cfg.Engine)
}
