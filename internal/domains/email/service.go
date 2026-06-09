package email

import (
	"bytes"
	"fmt"
	"net/smtp"

	"github.com/hsm-gustavo/authentication/internal/config"
)

type Service struct {
	auth smtp.Auth
	addr string
	from string
}

func NewService(cfg *config.SMTPConfig) *Service {
	return &Service{
		auth: smtp.PlainAuth(
			"",
			cfg.Username,
			cfg.Password,
			cfg.Host,
		),
		addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		from: cfg.From,
	}
}

func (s *Service) SendEmail(msg Message) error {
	data := s.buildMessage(msg)

	return smtp.SendMail(s.addr, s.auth, s.from, []string{msg.To}, data)
}

func (s *Service) buildMessage(msg Message) []byte {
	headers := map[string]string{
		"From":    fmt.Sprintf("Authentication <%s>", s.from),
		"To":      msg.To,
		"Subject": msg.Subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=\"UTF-8\"",
	}

	var body bytes.Buffer

	for key, value := range headers {
		body.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	body.WriteString("\r\n")
	body.WriteString(msg.Body)
	return body.Bytes()
}