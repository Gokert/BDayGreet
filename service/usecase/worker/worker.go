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
	utils "vk-rest/pkg"
	"vk-rest/pkg/models"
	"vk-rest/service/repository/profile"
)

type IWorker interface {
	StartWorker() error
}

type Worker struct {
	log      *logrus.Logger
	config   *models.MailConfigServer
	profiles profile.IProfileRepo
}

func GetWorker(psxCfg *configs.DbPsxConfig, log *logrus.Logger) (IWorker, error) {
	profileRepo, err := profile.GetPsxRepo(psxCfg, log)
	if err != nil {
		log.Error("Get GetFilmRepo error: ", err)
		return nil, err
	}

	port, err := strconv.Atoi(os.Getenv("PORT_HOST_MAIL"))
	if err != nil {
		log.Errorf("Error in GetPort: %v", err)
		return nil, err
	}

	cfg := &models.MailConfigServer{
		AddrEmail: os.Getenv("EMAIL_ADDRESS_SERVER"),
		Password:  os.Getenv("EMAIL_PASSWORD_SERVER"),
		Port:      port,
		AddrHost:  os.Getenv("ADDRESS_HOST_MAIL"),
	}

	worker := &Worker{
		log:      log,
		config:   cfg,
		profiles: profileRepo,
	}

	return worker, nil
}

func (w *Worker) StartWorker() error {
	w.log.Info("Worker started")
	c := cron.New()

	//запуск раз в день в 08.00
	err := c.AddFunc("0 0 8 * * *", w.HappyBirthday)
	if err != nil {
		return fmt.Errorf("cron error: %s", err.Error())
	}

	c.Start()
	return nil
}

func (w *Worker) HappyBirthday() {
	ctx := context.Background()

	employees, err := w.GetEmployeesBirthToday(ctx)
	if err != nil {
		w.log.Errorf("Error in CheckBirthday: %v", err)
		return
	}

	for _, employee := range employees {
		mail := &models.Mail{
			To:      employee.Email,
			Subject: utils.HeaderBirthdayEmp,
			Body:    utils.BodyBirthdayToEmp,
		}

		err = w.SendMail(mail, w.config)
		if err != nil {
			w.log.Info(err)
			return
		}

		employeesByBirthday, err := w.profiles.GetEmployeeByBirthday(ctx, employee.Id)
		if err != nil {
			w.log.Info(err)
			return
		}

		mail.Subject = utils.HeaderBirthdayEmp
		mail.Body = fmt.Sprintf(utils.BodyBirthdayFromEmp, employee.Login)

		for _, employeeByBirthday := range employeesByBirthday {
			mail.To = employeeByBirthday.Email

			err = w.SendMail(mail, w.config)
			if err != nil {
				w.log.Info(err)
				return
			}
		}

	}
}

func (w *Worker) GetEmployeesBirthToday(ctx context.Context) ([]*models.UserItem, error) {
	employees, err := w.profiles.GetBirthdayEmployees(ctx)
	if err != nil {
		return nil, err
	}

	return employees, nil
}

func (w *Worker) SendMail(mail *models.Mail, cfg *models.MailConfigServer) error {
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.AddrEmail)
	m.SetHeader("To", mail.To)
	m.SetHeader("Subject", mail.Subject)
	m.SetBody("text/plain", mail.Body)

	d := gomail.NewDialer(cfg.AddrHost, cfg.Port, cfg.AddrEmail, cfg.Password)

	err := d.DialAndSend(m)
	return err
}
