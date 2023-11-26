package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisCache struct {
	rdb *redis.Client
	ctx context.Context
}

func NewRedisCache() Cache {
	return &redisCache{}
}

func (r *redisCache) GetClient() (*CacheClient, error) {
	if r.rdb == nil {
		return nil, errors.New("no client")
	}
	return &CacheClient{IsCluster: false, Client: r.rdb}, nil
}

func (r *redisCache) Connect(uri string, password string, db int) error {
	ctx := context.TODO()
	rdb := redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	r.rdb = rdb
	r.ctx = ctx
	return nil
}

////////////////////////////////////////////////////

// func (r *redisCache) Ping() error {
// 	pong, err := r.rdb.Ping(r.ctx).Result()
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(pong, err)
// 	// Output: PONG <nil>

// 	return nil
// }

//PASS
func (r *redisCache) CacheByKey(key string, val interface{}, ex time.Duration) {
	if r.rdb == nil {
		return
	}

	j, _ := json.Marshal(val)
	r.rdb.Set(r.ctx, key, j, ex).Err()
}

//PASS
func (r *redisCache) GetByKey(key string) (string, error) {
	if r.rdb == nil {
		return "", errors.New("no redis client")
	}

	val, err := r.rdb.Get(r.ctx, key).Result()
	if err == redis.Nil || err != nil {
		return "", err
	}
	return val, nil

}

//PASS
func (r *redisCache) GetKeysByPattern(key string, count int64) ([]string, error) {

	if r.rdb == nil {
		return []string{}, errors.New("no redis client")
	}

	var cursor uint64
	result := []string{}
	for {
		var keys []string
		var err error
		keys, cursor, err = r.rdb.Scan(r.ctx, cursor, key, count).Result()

		if err != nil {
			return nil, err
		}
		result = append(result, keys...)
		if cursor == 0 {
			break
		}
	}
	return result, nil
}

//PASS
func (r *redisCache) DeleteKey(key string) {
	if r.rdb == nil {
		return
	}
	r.rdb.Del(r.ctx, key).Err()
}

//PASS
func (r *redisCache) BatchDeletionKeysByPattern(key string, count int64) {

	if r.rdb == nil {
		return
	}

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = r.rdb.Scan(r.ctx, cursor, key, count).Result()
		if err != nil {
			// panic(err)
			fmt.Println("batchdelete", err.Error())
			break
		}
		r.rdb.Del(r.ctx, keys...)
		if cursor == 0 {
			break
		}
	}

}

//PASS
func (r *redisCache) FlushDB() {
	if r.rdb == nil {
		return
	}
	r.rdb.FlushDB(r.ctx)

}

//PASS
func (r *redisCache) FlushAll() {
	r.rdb.FlushAll(r.ctx)
}
