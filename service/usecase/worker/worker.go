package worker

import (
	"context"
	"fmt"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"os"
	"strconv"
	"vk-rest/configs"
	"vk-rest/pkg/models"
	"vk-rest/service/repository/profile"
)

type IWorker interface {
	StartWorker() error
}

type Worker struct {
	log      *logrus.Logger
	profiles profile.IProfileRepo
}

func GetWorker(psxCfg *configs.DbPsxConfig, log *logrus.Logger) (IWorker, error) {
	profileRepo, err := profile.GetPsxRepo(psxCfg, log)
	if err != nil {
		log.Error("Get GetFilmRepo error: ", err)
		return nil, err
	}

	worker := &Worker{
		log:      log,
		profiles: profileRepo,
	}

	return worker, nil
}

func (w *Worker) StartWorker() error {
	w.log.Info("Worker started")

	c := cron.New()
	err := c.AddFunc("@every 1s", func() {
		ctx := context.Background()
		err := w.CheckBirthday(ctx)
		if err != nil {
			return
		}
	})
	if err != nil {
		return fmt.Errorf("cron error: %s", err.Error())
	}

	c.Start()

	return nil
}

func (w *Worker) CheckBirthday(ctx context.Context) error {
	employees, err := w.profiles.GetBirthdayEmployees(ctx)
	if err != nil {
		return err
	}

	for _, employee := range employees {
		mail := &models.Mail{
			To:      employee.Email,
			Subject: "Поздравление",
			Body:    "Поздравляем Вас с днём рождения!",
		}

		port, err := strconv.Atoi(os.Getenv("PORT_HOST_MAIL"))
		if err != nil {
			return err
		}

		cfg := &models.MailConfigServer{
			AddrEmail: os.Getenv("EMAIL_ADDRESS_SERVER"),
			Password:  os.Getenv("EMAIL_PASSWORD_SERVER"),
			Port:      port,
			AddrHost:  os.Getenv("ADDRESS_HOST_MAIL"),
		}

		err = w.SendMail(mail, cfg)
		if err != nil {
			w.log.Info(err)
			return err
		}

	}

	return nil
}

func (w *Worker) SendMail(mail *models.Mail, cfg *models.MailConfigServer) error {
	w.log.Info(cfg)

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.AddrEmail)
	m.SetHeader("To", mail.To)
	m.SetHeader("Subject", mail.Subject)
	m.SetBody("text/plain", mail.Body)

	d := gomail.NewDialer(cfg.AddrHost, cfg.Port, cfg.AddrEmail, cfg.Password)

	err := d.DialAndSend(m)
	return err
}
