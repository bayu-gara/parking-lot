package sugardb

import (
	"context"
	"errors"

	sugardblib "github.com/echovault/sugardb/sugardb"
	jsoniter "github.com/json-iterator/go"
)

type SugarDB struct {
	sugarDBClient *sugardblib.SugarDB
}

func Init() (*SugarDB, error) {
	sugarDBClient, err := sugardblib.NewSugarDB()
	if err != nil {
		return nil, err
	}

	return &SugarDB{sugarDBClient: sugarDBClient}, nil
}

func (sdb *SugarDB) Get(ctx context.Context, key string) (string, error) {
	return sdb.sugarDBClient.Get(key)
}

func (sdb *SugarDB) GetObj(ctx context.Context, key string, dest interface{}) error {
	jsonStr, err := sdb.Get(ctx, key)
	if err != nil {
		return err
	}

	err = jsoniter.UnmarshalFromString(jsonStr, dest)
	if err != nil {
		return err
	}

	return nil
}

func (sdb *SugarDB) MGet(ctx context.Context, keys ...string) (result []string, err error) {
	return sdb.sugarDBClient.MGet(keys...)
}

func (sdb *SugarDB) SetEX(ctx context.Context, key string, value string, expire int) error {
	_, ok, err := sdb.sugarDBClient.Set(key, value, sugardblib.SETOptions{ExpireOpt: sugardblib.SETEX, ExpireTime: expire})
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("Failed to set data to SugarDB")
	}

	return nil
}

func (sdb *SugarDB) SetEXObj(ctx context.Context, key string, value interface{}, expire int) error {
	jsonStr, err := jsoniter.MarshalToString(value)
	if err != nil {
		return errors.New("Failed to marshal json struct to string")
	}

	return sdb.SetEX(ctx, key, jsonStr, expire)
}

func (sdb *SugarDB) Delete(ctx context.Context, keys ...string) (int64, error) {
	totalDeletedData, err := sdb.sugarDBClient.Del(keys...)
	if err != nil {
		return 0, err
	}

	return int64(totalDeletedData), nil
}

func (sdb *SugarDB) SetNX(ctx context.Context, key string, value string, expire int) (bool, error) {
	_, ok, err := sdb.sugarDBClient.Set(key, value, sugardblib.SETOptions{ExpireOpt: sugardblib.SETEX, ExpireTime: expire, WriteOpt: sugardblib.SETNX})
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (sdb *SugarDB) LPush(ctx context.Context, key string, values ...string) (int64, error) {
	listNewLength, err := sdb.sugarDBClient.LPush(key, values...)
	if err != nil {
		return 0, err
	}

	return int64(listNewLength), nil
}

func (sdb *SugarDB) LPop(ctx context.Context, key string, count uint) ([]string, error) {
	return sdb.sugarDBClient.LPop(key, count)
}

func (sdb *SugarDB) RPush(ctx context.Context, key string, values ...string) (int64, error) {
	listNewLength, err := sdb.sugarDBClient.RPush(key, values...)
	if err != nil {
		return 0, err
	}

	return int64(listNewLength), nil
}

func (sdb *SugarDB) RPop(ctx context.Context, key string, count uint) ([]string, error) {
	return sdb.sugarDBClient.RPop(key, count)
}
