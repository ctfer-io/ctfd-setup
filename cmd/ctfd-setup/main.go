package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	ctfdsetup "github.com/ctfer-io/ctfd-setup"
	"github.com/ctfer-io/go-ctfd/api"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	Version    = "dev"
	Commit     = ""
	CommitDate = ""
	TreeState  = ""
)

func main() {
	app := &cli.App{
		Name:  "CTFd Setup",
		Usage: "Setup a CTFd instance from a fresh install or a resetted one.",
		Flags: []cli.Flag{
			cli.VersionFlag,
			cli.HelpFlag,
			&cli.StringFlag{
				Name:    "file",
				Usage:   "Configuration file to use for setting up CTFd. If let empty, will default the values and look for secrets in expected environment variables. For more info, refers to the documentation.",
				EnvVars: []string{"FILE", "PLUGIN_FILE"},
			},
			&cli.StringFlag{
				Name:     "url",
				Usage:    "URL to reach the CTFd instance.",
				Required: true,
				EnvVars:  []string{"CTFD_URL", "PLUGIN_CTFD_URL"},
			},
		},
		Action: run,
		Authors: []*cli.Author{
			{
				Name:  "Lucas Tesson - PandatiX",
				Email: "lucastesson@protonmail.com",
			},
		},
		Version: Version,
		Metadata: map[string]any{
			"version":   Version,
			"commit":    Commit,
			"date":      CommitDate,
			"treeState": TreeState,
		},
	}

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx *cli.Context) error {
	log := ctfdsetup.Log()

	// Read and unmarshal setup config file
	raw := []byte(ctfdsetup.DefaultConf)
	if ctx.IsSet("file") {
		f := ctx.String("file")
		log.Info("getting configuration file", zap.String("file", f))
		b, err := os.ReadFile(f)
		if err != nil {
			return errors.Wrapf(err, "reading file %s", f)
		}
		raw = b
	}

	var conf ctfdsetup.Conf
	if err := yaml.Unmarshal(raw, &conf); err != nil {
		return errors.Wrap(err, "unmarshalling configuration")
	}

	// Connect to CTFd
	url := ctx.String("url")
	log.Info("setting up CTFd instance", zap.String("url", url))
	nonce, session, err := api.GetNonceAndSession(url, api.WithContext(ctx.Context))
	if err != nil {
		return errors.Wrap(err, "getting CTFd nonce and session")
	}
	client := api.NewClient(url, nonce, session, "")

	// Flatten configuration and setup it
	if err := client.Setup(&api.SetupParams{
		CTFName:                conf.Name,
		CTFDescription:         conf.Description,
		UserMode:               conf.UserMode,
		ChallengeVisibility:    conf.Visibilities.Challenge,
		AccountVisibility:      conf.Visibilities.Account,
		ScoreVisibility:        conf.Visibilities.Score,
		RegistrationVisibility: conf.Visibilities.Registration,
		VerifyEmails:           conf.VerifyEmails,
		TeamSize:               conf.TeamSize,
		Name:                   conf.Admin.Name.ToString(),
		Email:                  conf.Admin.Email.ToString(),
		Password:               conf.Admin.Password.ToString(),
		CTFLogo:                conf.CTFLogo.ToInputFile(),
		CTFBanner:              conf.CTFBanner.ToInputFile(),
		CTFSmallIcon:           conf.CTFSmallIcon.ToInputFile(),
		CTFTheme:               conf.CTFTheme,
		ThemeColor:             conf.ThemeColor,
		Start:                  conf.Start,
		End:                    conf.End,
		Nonce:                  nonce,
	}, api.WithContext(ctx.Context)); err != nil {
		return errors.Wrap(err, "ctfd setup API call")
	}

	return nil
}
