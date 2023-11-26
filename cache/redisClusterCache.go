package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisClusterCache struct {
	rdb *redis.ClusterClient
	ctx context.Context
}

func NewRedisClusterCache() Cache {
	return &redisClusterCache{}
}

func (r *redisClusterCache) GetClient() (*CacheClient, error) {
	if r.rdb == nil {
		return nil, errors.New("no client")
	}
	return &CacheClient{IsCluster: true, ClusterClient: r.rdb}, nil
}

func (r *redisClusterCache) Connect(uri string, password string, db int) error {
	ctx := context.TODO()
	addr := []string{
		uri,
	}
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addr,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	r.rdb = rdb
	r.ctx = ctx
	return nil

}

// func (r *redisClusterCache) Ping() error {
// 	pong, err := r.rdb.Ping(r.ctx).Result()
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(pong, err)
// 	// Output: PONG <nil>

// 	return nil
// }

// //////////////////////////////////////////////////////
//PASS
func (r *redisClusterCache) CacheByKey(key string, val interface{}, ex time.Duration) {

	if r.rdb == nil {
		return
	}
	j, _ := json.Marshal(val)
	r.rdb.Set(r.ctx, key, j, ex).Err()

}

//PASS
func (r *redisClusterCache) GetByKey(key string) (string, error) {

	// fmt.Println("redis: ", r)
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
func (r *redisClusterCache) GetKeysByPattern(key string, count int64) ([]string, error) {

	if r.rdb == nil {
		return nil, errors.New("no redis client")
	}

	var allKeys []string
	queue := make(chan []string)
	errChan := make(chan error)
	done := make(chan bool)

	go func() {
		err := r.rdb.ForEachMaster(r.ctx, func(ctx context.Context, client *redis.Client) error {

			var cursor uint64
			for {
				var keys []string
				var err error
				keys, cursor, err = client.ScanType(ctx, cursor, key, count, "string").Result()

				if err != nil {
					fmt.Println("from callback", err.Error())
					return err
				}
				queue <- keys
				if cursor == 0 {
					break
				}
			}

			return nil

		})

		// after finish
		if err != nil {
			errChan <- err
		} else {
			done <- true
		}

	}()

	for {
		select {
		case err := <-errChan:
			return nil, err
		case <-done:
			return allKeys, nil
		case s := <-queue:
			allKeys = append(allKeys, s...)
		}
	}

}

func (r *redisClusterCache) BatchDeletionKeysByPattern(key string, count int64) {

	if r.rdb == nil {
		return
	}

	r.rdb.ForEachMaster(r.ctx, func(ctx context.Context, client *redis.Client) error {

		var cursor uint64
		for {
			var keys []string
			var err error
			keys, cursor, err = client.ScanType(ctx, cursor, key, count, "string").Result()

			if err != nil {
				fmt.Println("batchdelete", err.Error())
				return err
			}
			pipe := client.Pipeline()

			for _, k := range keys {
				pipe.Del(ctx, k)
			}
			pipe.Exec(ctx)

			if cursor == 0 {
				break
			}
		}

		return nil

	})

}

//PASS
func (r *redisClusterCache) DeleteKey(key string) {

	if r.rdb == nil {
		return
	}

	r.rdb.Del(r.ctx, key)
}

//must run on all masters
//PASS
func (r *redisClusterCache) FlushDB() {

	if r.rdb == nil {
		return
	}

	r.rdb.ForEachMaster(r.ctx, func(ctx context.Context, client *redis.Client) error {

		err := client.FlushDB(ctx).Err()
		if err != nil {
			return err
		}
		return nil

	})

}

//PASS
func (r *redisClusterCache) FlushAll() {

	if r.rdb == nil {
		return
	}

	r.rdb.ForEachMaster(r.ctx, func(ctx context.Context, client *redis.Client) error {

		err := client.FlushAll(ctx).Err()
		if err != nil {
			return err
		}
		return nil

	})
}
