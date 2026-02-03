package ctfdsetup_test

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var (
	CTFdURL string
)

const (
	CTFdURLvenv = "CTFD_URL"
)

var (
	envs []string
)

func TestMain(m *testing.M) {
	url, ok := os.LookupEnv(CTFdURLvenv)
	if !ok {
		fmt.Printf("Environment variable %s is not defined, please provide it", CTFdURLvenv)
		os.Exit(1)
	}
	CTFdURL = url

	// Prepare common environment variables, override GOCOVERDIR that go tests overrode
	pwd, _ := os.Getwd()
	covDir := filepath.Join(pwd, "coverdir")
	if err := os.MkdirAll(covDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	envs = append(os.Environ(),
		fmt.Sprintf("GOCOVERDIR=%s", covDir),
	)

	// Build the binary to avoid recompiling for each test
	cmd := exec.Command("go", "build", "-cover", "-o", "ctfd-setup", "cmd/ctfd-setup/main.go")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("err: %s ; output: %s", err, out)
	}
	defer func() {
		if err := os.Remove("./ctfd-setup"); err != nil {
			log.Fatal(err)
		}
	}()

	if sc := m.Run(); sc != 0 {
		log.Fatalf("Failed with status code %d", sc)
	}
}
