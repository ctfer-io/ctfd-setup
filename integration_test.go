package ctfdsetup_test

import (
	"bytes"
	"context"
	_ "embed"
	"os"
	"os/exec"
	"testing"

	ctfdsetup "github.com/ctfer-io/ctfd-setup"
	"github.com/ctfer-io/go-ctfd/api"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed examples/minimal/.ctfd.yaml
	minimalConf []byte
)

func Test_I_Minimal(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, reset(context.WithoutCancel(t.Context())))
	})

	// Check minimal configuration works fine
	conf := ctfdsetup.NewConfig()

	dec := yaml.NewDecoder(bytes.NewReader(minimalConf))
	dec.KnownFields(true)

	err := dec.Decode(conf)
	require.NoError(t, err)

	// Check it can be setup
	err = ctfdsetup.Setup(context.Background(), CTFdURL, "", conf)
	require.NoError(t, err)
}

func Test_I_NoFile(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, reset(context.WithoutCancel(t.Context())))
	})

	cmd := exec.CommandContext(t.Context(),
		"go", "run", "cmd/ctfd-setup/main.go",
		"--url", CTFdURL,
		"--appearance.name", "NoFileCTF",
		"--appearance.description", "A CTF configured with no file",
		"--admin.name", "ctfer",
		"--admin.email", "ctfer-io@protonmail.com",
		"--admin.password", "ctfer",
	)

	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func Test_I_NoBrackets2024(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, reset(context.WithoutCancel(t.Context())))
	})

	cmd := exec.CommandContext(t.Context(),
		"go", "run", "cmd/ctfd-setup/main.go",
		"--url", CTFdURL,
		"--directory", "examples/nobrackets2024",
		"--file", "examples/nobrackets2024/.ctfd.yaml",
	)

	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func Test_I_CLIOVerride(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, reset(context.WithoutCancel(t.Context())))
	})

	cmd := exec.CommandContext(t.Context(),
		"go", "run", "cmd/ctfd-setup/main.go",
		"--url", CTFdURL,
		"--file", "examples/cli-override/.ctfd.yaml",
		"--admin.name", "ctfer",
		"--admin.email", "ctfer-io@protonmail.com",
		"--admin.password", "ctfer",
	)

	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func Test_I_EnvOverride(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, reset(context.WithoutCancel(t.Context())))
	})

	cmd := exec.CommandContext(t.Context(),
		"go", "run", "cmd/ctfd-setup/main.go",
		"--url", CTFdURL,
		"--file", "examples/cli-override/.ctfd.yaml",
	)
	cmd.Env = append(os.Environ(),
		"ADMIN_NAME=ctfer",
		"ADMIN_EMAIL=ctfer-io@protonmail.com",
		"ADMIN_PASSWORD=ctfer",
	)

	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func reset(ctx context.Context) error {
	nonce, session, err := api.GetNonceAndSession(CTFdURL, api.WithContext(ctx))
	if err != nil {
		return err
	}

	client := api.NewClient(CTFdURL, nonce, session, "")

	if err := client.Login(&api.LoginParams{
		Name:     "ctfer",
		Password: "ctfer",
	}, api.WithContext(ctx)); err != nil {
		return err
	}
	return client.Reset(&api.ResetParams{
		Accounts:      ptr("y"),
		Submissions:   ptr("y"),
		Challenges:    ptr("y"),
		Pages:         ptr("y"),
		Notifications: ptr("y"),
	}, api.WithContext(ctx))
}

func ptr[T any](t T) *T {
	return &t
}
