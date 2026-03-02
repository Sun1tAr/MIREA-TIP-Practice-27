package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/tasks/internal/models"
)

type Client struct {
	rdb           *redis.Client
	logger        *logrus.Logger
	baseTTL       int
	jitterSeconds int
}

func NewClient(addr, password string, db int, baseTTL, jitterSeconds int, logger *logrus.Logger) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
	})

	return &Client{
		rdb:           rdb,
		logger:        logger,
		baseTTL:       baseTTL,
		jitterSeconds: jitterSeconds,
	}
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

// Ping проверяет доступность Redis
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// GetTask пытается получить задачу из кэша
func (c *Client) GetTask(ctx context.Context, id string) (*models.Task, error) {
	key := TaskKey(id)

	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		c.logger.WithField("key", key).Debug("CACHE MISS - key not found in Redis")
		return nil, nil
	}
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Warn("REDIS ERROR - failed to get key")
		return nil, err
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		c.logger.WithError(err).WithField("key", key).Error("CACHE ERROR - failed to unmarshal cached task")
		return nil, err
	}

	c.logger.WithField("key", key).Info("CACHE HIT - key found in Redis")
	return &task, nil
}

// SetTask сохраняет задачу в кэш
func (c *Client) SetTask(ctx context.Context, task *models.Task) error {
	key := TaskKey(task.ID)

	data, err := json.Marshal(task)
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("CACHE ERROR - failed to marshal task")
		return err
	}

	ttl := GetTTL(c.baseTTL, c.jitterSeconds)

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.WithError(err).WithField("key", key).Warn("REDIS ERROR - failed to set key")
		return err
	}

	c.logger.WithFields(logrus.Fields{
		"key": key,
		"ttl": ttl.Seconds(),
	}).Info("CACHE SET - key saved to Redis")
	return nil
}

// DeleteTask удаляет задачу из кэша
func (c *Client) DeleteTask(ctx context.Context, id string) error {
	key := TaskKey(id)

	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		c.logger.WithError(err).WithField("key", key).Warn("REDIS ERROR - failed to delete key")
		return err
	}

	c.logger.WithField("key", key).Debug("CACHE DELETE - key removed from Redis")
	return nil
}

// InvalidateList удаляет кэш списка задач
func (c *Client) InvalidateList(ctx context.Context) error {
	key := ListKey()

	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		c.logger.WithError(err).WithField("key", key).Warn("REDIS ERROR - failed to delete list key")
		return err
	}

	c.logger.WithField("key", key).Debug("CACHE DELETE - list key removed from Redis")
	return nil
}
