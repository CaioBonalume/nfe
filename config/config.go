package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type env struct {
	CERT_PATH string `env:"PFXPATH"`
	CERT_PASS string `env:"PFXPASSWORD"`
}

var Env env

func LoadEnv() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
	Env = env{
		CERT_PATH: os.Getenv("PFXPATH"),
		CERT_PASS: os.Getenv("PFXPASSWORD"),
	}
}
