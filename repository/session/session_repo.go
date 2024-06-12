package session

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"time"
	"vk-rest/configs"
	"vk-rest/pkg/models"
)

type ISessionRepo interface {
	AddSession(ctx context.Context, active models.Session) (bool, error)
	CheckActiveSession(ctx context.Context, sid string) (bool, error)
	GetUserLogin(ctx context.Context, sid string) (string, error)
	DeleteSession(ctx context.Context, sid string) (bool, error)
}

type SessionRepo struct {
	DB *redis.Client
}

func GetAuthRepo(cfg *configs.DbRedisCfg, log *logrus.Logger) (ISessionRepo, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Host,
		Password: cfg.Password,
		DB:       cfg.DbNumber,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Error("Ping redis error: ", err)
		return nil, err
	}

	log.Info("Redis created successful")
	return &SessionRepo{DB: redisClient}, nil
}

func (repo *SessionRepo) AddSession(ctx context.Context, active models.Session) (bool, error) {

	repo.DB.Set(ctx, active.SID, active.Login, 24*time.Hour)

	added, err := repo.CheckActiveSession(ctx, active.SID)
	if err != nil {
		return false, err
	}

	return added, nil
}

func (repo *SessionRepo) CheckActiveSession(ctx context.Context, sid string) (bool, error) {
	_, err := repo.DB.Get(ctx, sid).Result()
	if err == redis.Nil {
		return false, fmt.Errorf("key %s not found", sid)
	}

	if err != nil {
		return false, fmt.Errorf("get session id %s from redis failed", sid)
	}

	return true, err
}

func (repo *SessionRepo) GetUserLogin(ctx context.Context, sid string) (string, error) {
	value, err := repo.DB.Get(ctx, sid).Result()
	if err != nil {
		return "", fmt.Errorf("cannot find session %s", sid)
	}

	return value, nil
}

func (repo *SessionRepo) DeleteSession(ctx context.Context, sid string) (bool, error) {
	_, err := repo.DB.Del(ctx, sid).Result()
	if err != nil {
		return false, fmt.Errorf("cannot delete session %s", sid)
	}

	return true, nil
}
