package ctfdsetup_test

import (
	"fmt"
	"os"
	"testing"
)

var (
	CTFdURL string
)

const (
	CTFdURLvenv = "CTFD_URL"
)

func TestMain(m *testing.M) {
	url, ok := os.LookupEnv(CTFdURLvenv)
	if !ok {
		fmt.Printf("Environment variable %s is not defined, please provide it", CTFdURLvenv)
		os.Exit(1)
	}
	CTFdURL = url

	os.Exit(m.Run())
}
