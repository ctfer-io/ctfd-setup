package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	ctfdsetup "github.com/ctfer-io/ctfd-setup"
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

const (
	management    = "management"
	configuration = "configuration"
)

func main() {
	app := &cli.App{
		Name:  "CTFd Setup",
		Usage: "Setup (and update) a CTFd instance from a fresh install or an already-existing one.",
		Flags: []cli.Flag{
			cli.VersionFlag,
			cli.HelpFlag,
			&cli.StringFlag{
				Name:     "file",
				Value:    ".ctfd.yaml",
				Usage:    "Configuration file to use for setting up CTFd. If let empty, will default the values and look for secrets in expected environment variables. For more info, refers to the documentation.",
				EnvVars:  []string{"FILE", "PLUGIN_FILE"},
				Category: management,
			},
			&cli.StringFlag{
				Name:     "url",
				Usage:    "URL to reach the CTFd instance.",
				Required: true,
				EnvVars:  []string{"URL", "PLUGIN_URL"},
				Category: management,
			},
			// Configuration file
			// => global
			&cli.StringFlag{
				Name:     "global.name",
				Usage:    "The name of your CTF, displayed as is. Required.",
				EnvVars:  []string{"GLOBAL_NAME", "PLUGIN_GLOBAL_NAME"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "global.description",
				Usage:    "The description of your CTF, displayed as is. Required.",
				EnvVars:  []string{"GLOBAL_DESCRIPTION", "PLUGIN_GLOBAL_DESCRIPTION"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "global.mode",
				Usage:    "The mode of your CTFd, either users or teams.",
				Value:    "users",
				EnvVars:  []string{"GLOBAL_MODE", "PLUGIN_GLOBAL_MODE"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "global.team_size",
				Usage:    "The team size you want to enforce. Works only if global.mode is \"teams\".",
				EnvVars:  []string{"GLOBAL_TEAM_SIZE", "PLUGIN_GLOBAL_TEAM_SIZE"},
				Category: configuration,
			},
			// => visibilities
			&cli.StringFlag{
				Name:     "visibilities.challenge",
				Usage:    "The visibility for the challenges. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/)",
				Value:    "public",
				EnvVars:  []string{"VISIBILITIES_CHALLENGE", "PLUGIN_VISIBILITIES_CHALLENGE"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "visibilities.account",
				Usage:    "The visibility for the accounts. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/)",
				Value:    "public",
				EnvVars:  []string{"VISIBILITIES_ACCOUNT", "PLUGIN_VISIBILITIES_ACCOUNT"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "visibilities.score",
				Usage:    "The visibility for the scoreboard. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/)",
				Value:    "public",
				EnvVars:  []string{"VISIBILITIES_SCORE", "PLUGIN_VISIBILITIES_SCORE"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "visibilities.registration",
				Usage:    "The visibility for the registration. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/)",
				Value:    "public",
				EnvVars:  []string{"VISIBILITIES_REGISTRATION", "PLUGIN_VISIBILITIES_REGISTRATION"},
				Category: configuration,
			},
			// => front
			&cli.StringFlag{
				Name:     "front.theme",
				Usage:    "The frontend theme name.",
				Value:    "core",
				EnvVars:  []string{"FRONT_THEME", "PLUGIN_FRONT_THEME"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "front.theme_color",
				Usage:    "The frontend theme color.",
				EnvVars:  []string{"FRONT_THEME_COLOR", "PLUGIN_FRONT_THEME_COLOR"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "front.logo",
				Usage:    "The frontend logo. Provide a path to a locally-accessible file.",
				EnvVars:  []string{"FRONT_LOGO", "PLUGIN_FRONT_LOGO"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "front.banner",
				Usage:    "The frontend banner. Provide a path to a locally-accessible file.",
				EnvVars:  []string{"FRONT_BANNER", "PLUGIN_FRONT_BANNER"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "front.small_icon",
				Usage:    "The frontend small icon. Provide a path to a locally-accessible file.",
				EnvVars:  []string{"FRONT_SMALL_ICON", "PLUGIN_FRONT_SMALL_ICON"},
				Category: configuration,
			},
			// => admin
			&cli.StringFlag{
				Name:     "admin.name",
				Usage:    "The administrator name. Required.",
				EnvVars:  []string{"ADMIN_NAME", "PLUGIN_ADMIN_NAME"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "admin.email",
				Usage:    "The administrator email address. Required.",
				EnvVars:  []string{"ADMIN_EMAIL", "PLUGIN_ADMIN_EMAIL"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "admin.password",
				Usage:    "The administrator password. Required. Please use the environment variables.",
				EnvVars:  []string{"ADMIN_PASSWORD", "PLUGIN_ADMIN_PASSWORD"},
				Category: configuration,
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

	conf := &ctfdsetup.Config{
		Global: ctfdsetup.Global{
			Name:         ctx.String("global.name"),
			Description:  ctx.String("global.description"),
			Mode:         ctx.String("global.mode"),
			TeamSize:     toPtr(ctx.Int("global.team_size")),
			VerifyEmails: ctx.Bool("global.verify_emails"),
			Start:        ctx.String("global.start"),
			End:          ctx.String("global.end"),
		},
		Visibilities: ctfdsetup.Visibilities{
			Challenge:    ctx.String("visibilities.challenge"),
			Account:      ctx.String("visibilities.account"),
			Score:        ctx.String("visibilities.score"),
			Registration: ctx.String("visibilities.registration"),
		},
		Front: ctfdsetup.Front{
			Theme:      ctx.String("front.theme"),
			ThemeColor: ctx.String("front.color"),
			Logo:       toPtr(ctx.String("front.logo")),
			Banner:     toPtr(ctx.String("front.banner")),
			SmallIcon:  toPtr(ctx.String("front.small_icon")),
		},
		Admin: ctfdsetup.Admin{
			Name:     ctx.String("admin.name"),
			Email:    ctx.String("admin.email"),
			Password: ctx.String("admin.password"),
		},
	}

	// Read and unmarshal setup config file if any
	f := ctx.String("file")
	log.Info("getting configuration file", zap.String("file", f))
	if _, err := os.Stat(f); err == nil {
		log.Info("configuration file found", zap.String("file", f))
		b, err := os.ReadFile(f)
		if err != nil {
			return errors.Wrapf(err, "reading file %s", f)
		}
		if err := yaml.Unmarshal(b, &conf); err != nil {
			return errors.Wrap(err, "unmarshalling configuration")
		}
	}

	if err := conf.Validate(); err != nil {
		return err
	}

	// Connect to CTFd
	url := ctx.String("url")
	return ctfdsetup.Setup(ctx.Context, url, conf)
}

func toPtr[T comparable](t T) *T {
	if t == *new(T) {
		return &t
	}
	return nil
}
