package email

import (
	"fmt"
	"net/smtp"

	"github.com/hsm-gustavo/authentication/internal/config"
)

type Service struct {
	smtpConfig config.SMTPConfig
}


func NewService(smtpConfig config.SMTPConfig) *Service {
	return &Service{
		smtpConfig: smtpConfig,
	}
}

func (s *Service) SendEmail(to string, subject string, body string) error {
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n"
	fromHeader := fmt.Sprintf("From: %s\n\n", s.smtpConfig.From)

	message := []byte(subject + mime + fromHeader + body)

	auth := smtp.PlainAuth("", s.smtpConfig.Username, s.smtpConfig.Password, s.smtpConfig.Host)

	address := fmt.Sprintf("%s:%s", s.smtpConfig.Host, s.smtpConfig.Port)
	return smtp.SendMail(address, auth, s.smtpConfig.From, []string{to}, message)
}