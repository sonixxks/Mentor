package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"

	"github.com/joho/godotenv"
	"github.com/sonixxks/Mentor/backend/internal/config"
)

func main() {
	if err := godotenv.Load(); errors.Is(err, fs.ErrNotExist) {
		log.Println(".env не найден, используются переменные окружения")
	} else if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.App.Env)
}
