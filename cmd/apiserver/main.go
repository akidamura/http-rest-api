package main

import (
	"log"

	"github.com/akidamura/http-rest-api/internal/app/apiserver"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	config := apiserver.NewConfig()
	s := apiserver.New(config)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
