package apiserver

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../../.env")
	os.Setenv("DB_NAME", "restapi_test")

	os.Exit(m.Run())
}
