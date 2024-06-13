package profile

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/sirupsen/logrus"
	"time"
	"vk-rest/configs"
	utils "vk-rest/pkg"
	"vk-rest/pkg/models"
	pkg "vk-rest/pkg/sql"
)

type IProfileRepo interface {
	GetUser(ctx context.Context, login string, password []byte) (*models.UserItem, bool, error)
	FindUser(ctx context.Context, login string) (bool, error)
	CreateUser(ctx context.Context, user *models.SignupRequest, password []byte) error
	GetUserId(ctx context.Context, login string) (uint64, error)
	GetEmployees(ctx context.Context, offset, limit uint64) ([]*models.UserItem, error)
	GetBirthdayEmployees(ctx context.Context) ([]*models.UserItem, error)
}

type ProfileRepo struct {
	db *sql.DB
}

func GetPsxRepo(config *configs.DbPsxConfig, log *logrus.Logger) (*ProfileRepo, error) {
	fmt.Println(config)

	dsn := fmt.Sprintf("user=%s dbname=%s password= %s host=%s port=%d sslmode=%s",
		config.User, config.Dbname, config.Password, config.Host, config.Port, config.Sslmode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Errorf("sql open error: %s", err.Error())
		return nil, fmt.Errorf("get user r err: %s", err.Error())
	}

	r := &ProfileRepo{db: db}

	errs := make(chan error)
	go func() {
		errs <- r.pingDb(3, log)
	}()

	if err := <-errs; err != nil {
		log.Error(err.Error())
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	log.Info("Successfully connected to database")

	return r, nil
}

func (r *ProfileRepo) pingDb(timer uint32, log *logrus.Logger) error {
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

func (r *ProfileRepo) GetUser(ctx context.Context, login string, password []byte) (*models.UserItem, bool, error) {
	post := &models.UserItem{}

	err := r.db.QueryRowContext(ctx, pkg.GetUser, login, password).Scan(&post.Id, &post.Login, &post.Email, &post.Birthday)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get query user error: %s", err.Error())
	}

	return post, true, nil
}

func (r *ProfileRepo) FindUser(ctx context.Context, login string) (bool, error) {
	post := &models.UserItem{}

	err := r.db.QueryRowContext(ctx, pkg.FindUser, login).Scan(&post.Login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("find user query error: %s", err.Error())
	}

	return true, nil
}

func (r *ProfileRepo) CreateUser(ctx context.Context, user *models.SignupRequest, password []byte) error {
	var userID uint64
	err := r.db.QueryRowContext(ctx, pkg.CreateUser, user.Login, password, user.Email, user.Birthday).Scan(&userID)
	if err != nil {
		return fmt.Errorf("create user error: %s", err.Error())
	}

	return nil
}

func (r *ProfileRepo) GetUserId(ctx context.Context, login string) (uint64, error) {
	var userID uint64

	err := r.db.QueryRowContext(ctx, pkg.GetUserId, login).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("user not found for login: %s", login)
		}
		return 0, fmt.Errorf("get userpro file id error: %s", err.Error())
	}

	return userID, nil
}

func (r *ProfileRepo) GetEmployees(ctx context.Context, offset, limit uint64) ([]*models.UserItem, error) {
	users := make([]*models.UserItem, 0)

	rows, err := r.db.QueryContext(ctx, pkg.GetEmployees, offset, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users, nil
		}

		return nil, fmt.Errorf("get user query error: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		user := &models.UserItem{}

		err = rows.Scan(&user.Id, &user.Login, &user.Email, &user.Birthday)
		if err != nil {
			return nil, fmt.Errorf("get user rows scan error: %s", err.Error())
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *ProfileRepo) GetBirthdayEmployees(ctx context.Context) ([]*models.UserItem, error) {
	users := make([]*models.UserItem, 0)

	query := `
        SELECT id, login, email, birthday
        FROM profile
        WHERE EXTRACT(DAY FROM birthday) = EXTRACT(DAY FROM CURRENT_DATE)
        AND EXTRACT(MONTH FROM birthday) = EXTRACT(MONTH FROM CURRENT_DATE)
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.UserItem
		err := rows.Scan(&user.Id, &user.Login, &user.Email, &user.Birthday)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}
