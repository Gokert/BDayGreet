package usecase

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
	"vk-rest/configs"
	utils "vk-rest/pkg"
	"vk-rest/pkg/models"
	"vk-rest/repository/profile"
	"vk-rest/repository/session"
	"vk-rest/repository/sub"
)

type IProfileCore interface {
	GetUserId(ctx context.Context, sid string) (uint64, error)
	GetEmployees(ctx context.Context, limit, offset uint64) ([]*models.UserItem, error)
	CreateUserAccount(ctx context.Context, login string, password string) error
	FindUserAccount(ctx context.Context, login string, password string) (*models.UserItem, bool, error)
	FindUserByLogin(ctx context.Context, login string) (bool, error)
}

type ISessionCore interface {
	GetUserName(ctx context.Context, sid string) (string, error)
	CreateSession(ctx context.Context, login string) (models.Session, error)
	FindActiveSession(ctx context.Context, sid string) (bool, error)
	KillSession(ctx context.Context, sid string) error
}

type ISubCore interface {
	BirthdaySub(ctx context.Context, userId, subscriberId uint64) (bool, error)
	BirthdayUnSub(ctx context.Context, userId, subscriberId uint64) (bool, error)
}

type Core struct {
	log      *logrus.Logger
	profiles profile.IProfileRepo
	sessions session.ISessionRepo
	subs     sub.ISubRepo
}

func GetCore(psxCfg *configs.DbPsxConfig, redisCfg *configs.DbRedisCfg, log *logrus.Logger) (*Core, error) {
	profileRepo, err := profile.GetPsxRepo(psxCfg, log)
	if err != nil {
		log.Error("Get GetFilmRepo error: ", err)
		return nil, err
	}

	authRepo, err := session.GetAuthRepo(redisCfg, log)
	if err != nil {
		log.Error("Get GetAuthRepo error: ", err)
		return nil, err
	}

	subsRepo, err := sub.GetSubRepo(psxCfg, log)
	if err != nil {
		log.Error("Get GetSubRepo error: ", err)
		return nil, err
	}

	core := &Core{
		log:      log,
		sessions: authRepo,
		profiles: profileRepo,
		subs:     subsRepo,
	}

	return core, nil
}

func (c *Core) GetUserId(ctx context.Context, sid string) (uint64, error) {
	login, err := c.sessions.GetUserLogin(ctx, sid)

	if err != nil {
		c.log.Errorf("get user login error: %s", err.Error())
		return 0, fmt.Errorf("get user login error: %s", err.Error())
	}

	id, err := c.profiles.GetUserId(ctx, login)
	if err != nil {
		c.log.Errorf("get user id error: %s", err.Error())
		return 0, fmt.Errorf("get user id error: %s", err.Error())
	}

	return id, nil
}

func (c *Core) GetUserName(ctx context.Context, sid string) (string, error) {
	login, err := c.sessions.GetUserLogin(ctx, sid)

	if err != nil {
		c.log.Errorf("get user name error: %s", err.Error())
		return "", fmt.Errorf("get user name error: %s", err.Error())
	}

	return login, nil
}

func (c *Core) CreateSession(ctx context.Context, login string) (models.Session, error) {
	sid := utils.RandStringRunes(32)

	newSession := models.Session{
		Login:     login,
		SID:       sid,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	sessionAdded, err := c.sessions.AddSession(ctx, newSession)

	if !sessionAdded && err != nil {
		return models.Session{}, err
	}

	if !sessionAdded {
		return models.Session{}, nil
	}

	return newSession, nil
}

func (c *Core) FindActiveSession(ctx context.Context, sid string) (bool, error) {
	login, err := c.sessions.CheckActiveSession(ctx, sid)

	if err != nil {
		c.log.Errorf("find active session error: %s", err.Error())
		return false, fmt.Errorf("find active session error: %s", err.Error())
	}

	return login, nil
}

func (c *Core) KillSession(ctx context.Context, sid string) error {
	_, err := c.sessions.DeleteSession(ctx, sid)

	if err != nil {
		c.log.Errorf("delete session error: %s", err.Error())
		return fmt.Errorf("delete sessionerror: %s", err.Error())
	}

	return nil
}

func (c *Core) CreateUserAccount(ctx context.Context, login string, password string) error {
	hashPassword := utils.HashPassword(password)
	err := c.profiles.CreateUser(ctx, login, hashPassword)
	if err != nil {
		c.log.Errorf("create user account error: %s", err.Error())
		return fmt.Errorf("create user account error: %s", err.Error())
	}

	return nil
}

func (c *Core) FindUserAccount(ctx context.Context, login string, password string) (*models.UserItem, bool, error) {
	hashPassword := utils.HashPassword(password)
	user, found, err := c.profiles.GetUser(ctx, login, hashPassword)
	if err != nil {
		c.log.Errorf("find user error: %s", err.Error())
		return nil, false, fmt.Errorf("find user account error: %s", err.Error())
	}
	return user, found, nil
}

func (c *Core) FindUserByLogin(ctx context.Context, login string) (bool, error) {
	found, err := c.profiles.FindUser(ctx, login)
	if err != nil {
		c.log.Errorf("find user by login error: %s", err.Error())
		return false, fmt.Errorf("find user by login error: %s", err.Error())
	}

	return found, nil
}

func (c *Core) GetEmployees(ctx context.Context, limit, offset uint64) ([]*models.UserItem, error) {
	users, err := c.profiles.GetEmployees(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Core) BirthdaySub(ctx context.Context, userId, subscriberId uint64) (bool, error) {
	res, err := c.subs.BirthdaySub(ctx, userId, subscriberId)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (c *Core) BirthdayUnSub(ctx context.Context, userId, subscriberId uint64) (bool, error) {
	res, err := c.subs.BirthdayUnSub(ctx, userId, subscriberId)
	if err != nil {
		return false, err
	}

	return res, nil
}
