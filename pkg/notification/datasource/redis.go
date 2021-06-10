package datasource

import (
	redis "github.com/go-redis/redis/v7"
)

const (
	RegistryKey = "noti_reg"
	QueueKey    = "noti_queue"
)

type RedisDataSource struct {
	client *redis.Client
}

func NewRedisDataSource(addr string, password string, db int) (*RedisDataSource, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := c.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &RedisDataSource{
		client: c,
	}, nil
}

func (s *RedisDataSource) Save(id string, namespace string, data []byte) error {
	return s.client.HSet(RegistryKey, id + namespace, data).Err()
}

func (s *RedisDataSource) Load(id string, namespace string) ([]byte, error) {
	r := s.client.HGet(RegistryKey, id + namespace)
	if r.Err() != nil {
		return nil, r.Err()
	}

	return []byte(r.Val()), nil
}

func (s *RedisDataSource) Push(data []byte) error {
	return s.client.LPush(QueueKey, data).Err()
}

func (s *RedisDataSource) Pop() ([]byte, error) {
	r := s.client.RPop(QueueKey)
	if r.Err() != nil {
		return nil, r.Err()
	}

	return []byte(r.Val()), nil
}
