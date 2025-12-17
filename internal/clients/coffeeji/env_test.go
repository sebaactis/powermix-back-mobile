package coffeeji

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {

	if err := godotenv.Load("../../../.env"); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}