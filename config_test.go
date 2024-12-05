package ctfdsetup_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	ctfdsetup "github.com/ctfer-io/ctfd-setup"
	"github.com/ctfer-io/go-ctfd/api"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const (
	dir = "examples"
)

func Test_U_ConfigSchema(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	cfg := ctfdsetup.NewConfig()

	schema, err := cfg.Schema()
	assert.NoError(err)
	assert.NotEmpty(schema)
}

func Test_F_Examples(t *testing.T) {
	url, ok := os.LookupEnv("URL")
	if !ok {
		t.Fatal("environment variable URL is not defined")
	}

	files, err := os.ReadDir(dir)
	if !assert.NoError(t, err) {
		return
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()

			c, err := os.ReadFile(filepath.Join(dir, f.Name()))
			if !assert.NoError(err) {
				return
			}

			// Extract config
			cfg := ctfdsetup.NewConfig()
			err = yaml.Unmarshal(c, cfg)
			assert.NoError(err)

			err = cfg.Validate()
			if !assert.NoError(err) {
				return
			}

			// Login then reset (required to run multiple test cases)
			defer func() {
				nonce, session, err := api.GetNonceAndSession(url, api.WithContext(ctx))
				assert.NoError(err)
				client := api.NewClient(url, nonce, session, "")

				err = client.Login(&api.LoginParams{
					Name:     cfg.Admin.Name,
					Password: cfg.Admin.Password.Content,
				}, api.WithContext(ctx))
				assert.NoError(err)

				err = client.Reset(&api.ResetParams{
					Accounts:      ptr("true"),
					Submissions:   ptr("true"),
					Challenges:    ptr("true"),
					Pages:         ptr("true"),
					Notifications: ptr("true"),
				})
				assert.NoError(err)
			}()

			// Setup CTFd
			err = ctfdsetup.Setup(ctx, url, "", cfg)
			assert.NoError(err)
		})
	}
}

func ptr[T any](t T) *T {
	return &t
}
