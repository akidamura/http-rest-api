package main

import (
	"log"

	"github.com/akidamura/http-rest-api/internal/app/apiserver"
)

func main() {
	config := apiserver.NewConfig()
	s := apiserver.New(config)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
