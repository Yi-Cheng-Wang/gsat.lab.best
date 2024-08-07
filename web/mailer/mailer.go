package mailer

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"path/filepath"
	"strings"
)

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unknown from server")
		}
	}
	return nil, nil
}

type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

// GetSMTPConfig retrieves the SMTP configuration from the database
func GetSMTPConfig(db *sql.DB) (SMTPConfig, error) {
	var content string
	err := db.QueryRow("SELECT content FROM system_secrets WHERE name = ?", "smtp_config").Scan(&content)
	if err != nil {
		return SMTPConfig{}, err
	}

	return parseSMTPConfig(content), nil
}

// parseSMTPConfig parses the content string into an SMTPConfig struct
func parseSMTPConfig(content string) SMTPConfig {
	parts := strings.Split(content, ";")
	config := SMTPConfig{}
	for _, part := range parts {
		keyValue := strings.SplitN(part, ":", 2)
		if len(keyValue) != 2 {
			continue
		}
		key, value := keyValue[0], keyValue[1]
		switch key {
		case "host":
			config.Host = value
		case "port":
			config.Port = value
		case "user":
			config.User = value
		case "password":
			config.Password = value
		}
	}
	return config
}

// SendEmail sends an email using the specified template
func SendEmail(db *sql.DB, to, subject, templateName string, data interface{}) error {
	config, err := GetSMTPConfig(db)
	if err != nil {
		return err
	}

	// Connect to the SMTP server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", config.Host, config.Port))
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return err
	}
	defer c.Quit()

	// Start TLS
	tlsconfig := &tls.Config{
		ServerName: config.Host,
	}

	if err = c.StartTLS(tlsconfig); err != nil {
		return err
	}

	// Authenticate
	auth := LoginAuth(config.User, config.Password)

	if err = c.Auth(auth); err != nil {
		return err
	}

	// Set the sender and recipient
	if err = c.Mail(config.User); err != nil {
		return err
	}

	if err = c.Rcpt(to); err != nil {
		return err
	}

	// Send the email body
	wc, err := c.Data()
	if err != nil {
		return err
	}

	// Load and parse the email template
	tmpl, err := template.ParseFiles(filepath.Join("tem", templateName))
	if err != nil {
		return err
	}

	// Create the email body
	var body strings.Builder
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0\r\n"+
		"Content-Type: text/html; charset=utf-8\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body.String()))

	_, err = wc.Write(msg)
	if err != nil {
		return err
	}

	err = wc.Close()
	if err != nil {
		return err
	}

	return nil
}
