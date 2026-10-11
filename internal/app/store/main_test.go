package store_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../../.env")   // подгружаем .env
	os.Setenv("DB_NAME", "restapi_test") // переопределяем базу на тестовую

	os.Exit(m.Run())
}
