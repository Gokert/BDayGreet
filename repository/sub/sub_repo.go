package sub

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
	"vk-rest/configs"
	utils "vk-rest/pkg"
	errs "vk-rest/pkg/errors"
)

type ISubRepo interface {
	BirthdaySub(ctx context.Context, userId, subscriberId uint64) (bool, error)
	BirthdayUnSub(ctx context.Context, userId, subscriberId uint64) (bool, error)
}

type SubRepo struct {
	db *sql.DB
}

func GetSubRepo(config *configs.DbPsxConfig, log *logrus.Logger) (*SubRepo, error) {
	dsn := fmt.Sprintf("user=%s dbname=%s password= %s host=%s port=%d sslmode=%s",
		config.User, config.Dbname, config.Password, config.Host, config.Port, config.Sslmode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Errorf("sql open error: %s", err.Error())
		return nil, fmt.Errorf("get user repo err: %s", err.Error())
	}

	repo := &SubRepo{db: db}

	errs := make(chan error)
	go func() {
		errs <- repo.pingDb(3, log)
	}()

	if err := <-errs; err != nil {
		log.Error(err.Error())
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	log.Info("Successfully connected to database")

	return repo, nil
}

func (r *SubRepo) pingDb(timer uint32, log *logrus.Logger) error {
	var err error
	var retries int

	for retries < utils.MaxRetries {
		err = r.db.Ping()
		if err == nil {
			return nil
		}

		retries++
		log.Errorf("sql ping error: %s", err.Error())
		time.Sleep(time.Duration(timer) * time.Second)
	}

	return fmt.Errorf("sql max pinging error: %s", err.Error())
}

func (r *SubRepo) BirthdaySub(ctx context.Context, userId, subscriberId uint64) (bool, error) {
	fmt.Printf("%d %d", userId, subscriberId)

	res, err := r.db.ExecContext(ctx, "INSERT INTO subscriber (id_subscribe_from, id_subscribe_to) VALUES ($1, $2)", userId, subscriberId)
	if err != nil {
		if err.Error() == errs.ErrDuplicateSub.Error() {
			return false, err
		}

		return false, fmt.Errorf("insert subscriber err: %s", err.Error())
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %s", err.Error())
	}

	if rowsAffected == 0 {
		return false, errs.ErrNotFound
	}

	return true, nil
}

func (r *SubRepo) BirthdayUnSub(ctx context.Context, userId, subscriberId uint64) (bool, error) {
	res, err := r.db.ExecContext(ctx, "DELETE FROM subscriber WHERE id_subscribe_from = $1 AND id_subscribe_to = $2", userId, subscriberId)
	if err != nil {
		return false, fmt.Errorf("insert subscriber err: %s", err.Error())
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %s", err.Error())
	}

	if rowsAffected == 0 {
		return false, errs.ErrNotFound
	}

	return true, nil
}
