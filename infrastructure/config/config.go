package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL   string
	HTTPAddr      string
	AuthSecret    string
	AdminToken    string
	AdminID       string
	AllowedOrigin string
	ImageBase     string
	LineLoginID   string
	LineSecret    string
	LineToken     string
	LineRecipient string
}

func Load() (Config, error) {
	c := Config{DatabaseURL: os.Getenv("DATABASE_URL"), HTTPAddr: os.Getenv("HTTP_ADDR"), AuthSecret: os.Getenv("AUTH_SECRET"), AdminToken: os.Getenv("ADMIN_API_TOKEN"), AdminID: os.Getenv("ADMIN_USER_ID"), AllowedOrigin: os.Getenv("FRONTEND_ORIGIN"), ImageBase: os.Getenv("PRODUCT_IMAGE_BASE_URL"), LineLoginID: os.Getenv("LINE_LOGIN_CHANNEL_ID"), LineSecret: os.Getenv("LINE_CHANNEL_SECRET")}
	if c.DatabaseURL == "" {
		return c, errors.New("DATABASE_URL is required")
	}
	c.LineToken = os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	c.LineRecipient = os.Getenv("LINE_ADMIN_USER_ID")
	if c.HTTPAddr == "" {
		c.HTTPAddr = "127.0.0.1:8080"
	}
	return c, nil
}
