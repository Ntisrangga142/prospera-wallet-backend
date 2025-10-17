package utils

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
)

func SendResetPasswordEmail(to string, resetURL string, typeReset string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("Prospera <%s>", os.Getenv("SMTP_EMAIL")))
	m.SetHeader("To", to)
	m.SetHeader("Subject", fmt.Sprintf("Reset %s", typeReset))
	m.SetBody("text/html", fmt.Sprintf("<p>Click link berikut untuk reset %s:</p><a href='%s'>Reset %s</a>", typeReset, resetURL, typeReset))

	d := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("SMTP_EMAIL"), os.Getenv("SMTP_PASSWORD"))

	return d.DialAndSend(m)
}

func GenerateRandomToken() string {
	return uuid.New().String()
}
