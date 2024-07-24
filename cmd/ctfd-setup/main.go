package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
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
			&cli.StringFlag{
				Name:     "api_key",
				Usage:    "The API key to use (for instance for a CI SA), used for updating a running CTFd instance.",
				EnvVars:  []string{"API_KEY", "PLUGIN_API_KEY"},
				Category: management,
			},
			// Configuration file
			// => Appearance
			&cli.StringFlag{
				Name:     "appearance.name",
				Usage:    "The name of your CTF, displayed as is.",
				EnvVars:  []string{"APPEARANCE_NAME", "PLUGIN_APPEARANCE_NAME"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "appearance.description",
				Usage:    "The description of your CTF, displayed as is.",
				EnvVars:  []string{"APPEARANCE_DESCRIPTION", "PLUGIN_APPEARANCE_DESCRIPTION"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "appearance.default_locale",
				Usage:    "The default language for the users.",
				EnvVars:  []string{"APPEARANCE_DEFAULT_LOCALE", "PLUGIN_APPEARANCE_DEFAULT_LOCALE"},
				Category: configuration,
			},
			// => Theme
			&cli.StringFlag{
				Name:     "theme.logo",
				Usage:    "The frontend logo. Provide a path to a locally-accessible file.",
				EnvVars:  []string{"THEME_LOGO", "PLUGIN_THEME_LOGO"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.small_icon",
				Usage:    "The frontend small icon. Provide a path to a locally-accessible file.",
				EnvVars:  []string{"THEME_SMALL_ICON", "PLUGIN_THEME_SMALL_ICON"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.name",
				Usage:    "The frontend theme name.",
				Value:    "core",
				EnvVars:  []string{"THEME_NAME", "PLUGIN_THEME_NAME"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.color",
				Usage:    "The frontend theme color.",
				EnvVars:  []string{"THEME_COLOR", "PLUGIN_THEME_COLOR"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.header",
				Usage:    "The frontend header. Provide a path to a locally-accessible file.",
				EnvVars:  []string{"THEME_HEADER", "PLUGIN_THEME_HEADER"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.footer",
				Usage:    "The frontend footer. Provide a path to a locally-accessible file.",
				EnvVars:  []string{"THEME_FOOTER", "PLUGIN_THEME_FOOTER"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.settings",
				Usage:    "The frontend settings (JSON). Provide a path to a locally-accessible file.",
				EnvVars:  []string{"THEME_SETTINGS", "PLUGIN_THEME_SETTINGS"},
				Category: configuration,
			},
			// => Accounts
			&cli.StringFlag{
				Name:     "accounts.domain_whitelist",
				Usage:    "The domain whitelist (a list separated by colons) to allow users to have email addresses from.",
				EnvVars:  []string{"ACCOUNTS_DOMAIN_WHITELIST", "PLUGIN_ACCOUNTS_DOMAIN_WHITELIST"},
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "accounts.verify_emails",
				Usage:    "Whether to verify emails once a user register or not.",
				Value:    false,
				EnvVars:  []string{"ACCOUNTS_VERIFY_EMAILS", "PLUGIN_ACCOUNTS_VERIFY_EMAILS"},
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "accounts.team_creation",
				Usage:    "Whether to allow team creation by players or not.",
				EnvVars:  []string{"ACCOUNTS_TEAM_CREATION", "PLUGIN_ACCOUNTS_TEAM_CREATION"},
				Category: configuration,
			},
			&cli.IntFlag{
				Name:     "accounts.team_size",
				Usage:    "Maximum size (number of players) in a team.",
				EnvVars:  []string{"ACCOUNTS_TEAM_SIZE", "PLUGIN_ACCOUNTS_TEAM_SIZE"},
				Category: configuration,
			},
			&cli.IntFlag{
				Name:     "accounts.num_teams",
				Usage:    "The total number of teams allowed.",
				EnvVars:  []string{"ACCOUNTS_NUM_TEAMS", "PLUGIN_ACCOUNTS_NUM_TEAMS"},
				Category: configuration,
			},
			&cli.IntFlag{
				Name:     "accounts.num_users",
				Usage:    "The total number of users allowed.",
				EnvVars:  []string{"ACCOUNTS_NUM_USERS", "PLUGIN_ACCOUNTS_NUM_USERS"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "accounts.team_disbanding",
				Usage:    "Whether to allow teams to be disbanded or not. Could be inactive_only or disabled.",
				EnvVars:  []string{"ACCOUNTS_TEAM_DISBANDING", "PLUGIN_ACCOUNTS_TEAM_DISBANDING"},
				Category: configuration,
			},
			&cli.IntFlag{
				Name:     "accounts.incorrect_submissions_per_minute",
				Usage:    "Maximum number of invalid submissions per minute (per user/team). We suggest you use it as part of an anti-brute-force strategy (rate limiting).",
				EnvVars:  []string{"ACCOUNTS_INCORRECT_SUBMISSIONS_PER_MINUTE", "PLUGIN_ACCOUNTS_INCORRECT_SUBMISSIONS_PER_MINUTE"},
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "accounts.name_changes",
				Usage:    "Whether a user can change its name or not.",
				EnvVars:  []string{"ACCOUNTS_NAME_CHANGES", "PLUGIN_ACCOUNTS_NAME_CHANGES"},
				Category: configuration,
			},
			// => Pages
			&cli.StringFlag{
				Name:     "pages.robots_txt",
				Usage:    "Define the /robots.txt file content, for web crawlers indexing.",
				EnvVars:  []string{"PAGES_ROBOTS_TXT", "PLUGIN_PAGES_ROBOTS_TXT"},
				Category: configuration,
			},
			// => MajorLeagueCyber
			&cli.StringFlag{
				Name:     "major_league_cyber.client_id",
				Usage:    "The MajorLeagueCyber OAuth ClientID.",
				EnvVars:  []string{"MAJOR_LEAGUE_CYBER_CLIENT_ID", "PLUGIN_MAJOR_LEAGUE_CYBER_CLIENT_ID"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "major_league_cyber.client_secret",
				Usage:    "The MajorLeagueCyber OAuth Client Secret.",
				EnvVars:  []string{"MAJOR_LEAGUE_CYBER_CLIENT_SECRET", "PLUGIN_MAJOR_LEAGUE_CYBER_CLIENT_SECRET"},
				Category: configuration,
			},
			// => Settings
			&cli.StringFlag{
				Name:     "settings.challenge_visibility",
				Usage:    "The visibility for the challenges. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).",
				Value:    "public",
				EnvVars:  []string{"SETTINGS_CHALLENGE_VISIBILITY", "PLUGIN_SETTINGS_CHALLENGE_VISIBILITY"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "settings.account_visibility",
				Usage:    "The visibility for the accounts. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).",
				Value:    "public",
				EnvVars:  []string{"SETTINGS_ACCOUNT_VISIBILITY", "PLUGIN_SETTINGS_ACCOUNT_VISIBILITY"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "settings.score_visibility",
				Usage:    "The visibility for the scoreboard. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).",
				Value:    "public",
				EnvVars:  []string{"SETTINGS_SCORE_VISIBILITY", "PLUGIN_SETTINGS_SCORE_VISIBILITY"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "settings.registration_visibility",
				Usage:    "The visibility for the registration. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).",
				Value:    "public",
				EnvVars:  []string{"SETTINGS_REGISTRATION_VISIBILITY", "PLUGIN_SETTINGS_REGISTRATION_VISIBILITY"},
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "settings.paused",
				Usage:    "Whether the CTFd is paused or not.",
				EnvVars:  []string{"SETTINGS_PAUSED", "PLUGIN_SETTINGS_PAUSED"},
				Category: configuration,
			},
			// => Security
			&cli.BoolFlag{
				Name:     "security.html_sanitization",
				Usage:    "Whether to turn on HTML sanitization or not.",
				EnvVars:  []string{"SECURITY_HTML_SANITIZATION", "PLUGIN_SECURITY_HTML_SANITIZATION"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "security.registration_code",
				Usage:    "The registration code (secret) to join the CTF.",
				EnvVars:  []string{"SECURITY_REGISTRATION_CODE", "PLUGIN_SECURITY_REGISTRATION_CODE"},
				Category: configuration,
			},
			// => Email
			&cli.StringFlag{
				Name:     "email.registration.subject",
				Usage:    "The email registration subject of the mail.",
				EnvVars:  []string{"EMAIL_REGISTRATION_SUBJECT", "PLUGIN_EMAIL_REGISTRATION_SUBJECT"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.registration.body",
				Usage:    "The email registration body of the mail.",
				EnvVars:  []string{"EMAIL_REGISTRATION_BODY", "PLUGIN_EMAIL_REGISTRATION_BODY"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.confirmation.subject",
				Usage:    "The email confirmation subject of the mail.",
				EnvVars:  []string{"EMAIL_CONFIRMATION_SUBJECT", "PLUGIN_EMAIL_CONFIRMATION_SUBJECT"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.confirmation.body",
				Usage:    "The email confirmation body of the mail.",
				EnvVars:  []string{"EMAIL_CONFIRMATION_BODY", "PLUGIN_EMAIL_CONFIRMATION_BODY"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.new_account.subject",
				Usage:    "The email new_account subject of the mail.",
				EnvVars:  []string{"EMAIL_NEW_ACCOUNT_SUBJECT", "PLUGIN_EMAIL_NEW_ACCOUNT_SUBJECT"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.new_account.body",
				Usage:    "The email new_account body of the mail.",
				EnvVars:  []string{"EMAIL_NEW_ACCOUNT_BODY", "PLUGIN_EMAIL_NEW_ACCOUNT_BODY"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.password_reset.subject",
				Usage:    "The email password_reset subject of the mail.",
				EnvVars:  []string{"EMAIL_PASSWORD_RESET_SUBJECT", "PLUGIN_EMAIL_PASSWORD_RESET_SUBJECT"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.password_reset.body",
				Usage:    "The email password_reset body of the mail.",
				EnvVars:  []string{"EMAIL_PASSWORD_RESET_BODY", "PLUGIN_EMAIL_PASSWORD_RESET_BODY"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.password_reset_confirmation.subject",
				Usage:    "The email password_reset_confirmation subject of the mail.",
				EnvVars:  []string{"EMAIL_PASSWORD_RESET_CONFIRMATION_SUBJECT", "PLUGIN_EMAIL_PASSWORD_RESET_CONFIRMATION_SUBJECT"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.password_reset_confirmation.body",
				Usage:    "The email password_reset_confirmation body of the mail.",
				EnvVars:  []string{"EMAIL_PASSWORD_RESET_CONFIRMATION_BODY", "PLUGIN_EMAIL_PASSWORD_RESET_CONFIRMATION_BODY"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.from",
				Usage:    "The 'From:' to sent to mail with.",
				EnvVars:  []string{"EMAIL_MAIL_FROM", "PLUGIN_EMAIL_MAIL_FROM"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.server",
				Usage:    "The mail server to use.",
				EnvVars:  []string{"EMAIL_MAIL_SERVER", "PLUGIN_EMAIL_MAIL_SERVER"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.port",
				Usage:    "The mail server port to reach.",
				EnvVars:  []string{"EMAIL_MAIL_SERVER_PORT", "PLUGIN_EMAIL_MAIL_SERVER_PORT"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.username",
				Usage:    "The username to log in to the mail server.",
				EnvVars:  []string{"EMAIL_USERNAME", "PLUGIN_EMAIL_USERNAME"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.password",
				Usage:    "The password to log in to the mail server.",
				EnvVars:  []string{"EMAIL_PASSWORD", "PLUGIN_EMAIL_PASSWORD"},
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "email.tls_ssl",
				Usage:    "Whether to turn on TLS/SSL or not.",
				EnvVars:  []string{"EMAIL_TLS_SSL", "PLUGIN_EMAIL_TLS_SSL"},
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "email.starttls",
				Usage:    "Whether to turn on STARTTLS or not.",
				EnvVars:  []string{"EMAIL_STARTTLS", "PLUGIN_EMAIL_STARTTLS"},
				Category: configuration,
			},
			// => Time
			&cli.StringFlag{
				Name:     "time.start",
				Usage:    "The start timestamp at which the CTFd will open.",
				EnvVars:  []string{"TIME_START", "PLUGIN_TIME_START"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "time.end",
				Usage:    "The end timestamp at which the CTFd will close.",
				EnvVars:  []string{"TIME_END", "PLUGIN_TIME_END"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "time.freeze",
				Usage:    "The freeze timestamp at which the CTFd will remain open but won't accept any further submissions.",
				EnvVars:  []string{"TIME_FREEZE", "PLUGIN_TIME_FREEZE"},
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "time.view_after",
				Usage:    "Whether allows users to view challenges after end or not.",
				EnvVars:  []string{"TIME_VIEW_AFTER", "PLUGIN_TIME_VIEW_AFTER"},
				Category: configuration,
			},
			// => Social
			&cli.BoolFlag{
				Name:     "social.shares",
				Usage:    "Whether to enable users share they solved a challenge or not.",
				EnvVars:  []string{"SOCIAL_SHARES", "PLUGIN_SOCIAL_SHARES"},
				Category: configuration,
			},
			// => Legal
			&cli.StringFlag{
				Name:     "legal.tos.url",
				Usage:    "The Terms of Services URL.",
				EnvVars:  []string{"LEGAL_TOS_URL", "PLUGIN_LEGAL_TOS_URL"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "legal.tos.content",
				Usage:    "The Terms of Services content.",
				EnvVars:  []string{"LEGAL_TOS_CONTENT", "PLUGIN_LEGAL_TOS_CONTENT"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "legal.privacy_policy.url",
				Usage:    "The Privacy Policy URL.",
				EnvVars:  []string{"LEGAL_PRIVACY_POLICY_URL", "PLUGIN_LEGAL_PRIVACY_POLICY_URL"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "legal.privacy_policy.content",
				Usage:    "The Privacy Policy content.",
				EnvVars:  []string{"LEGAL_PRIVACY_POLICY_CONTENT", "PLUGIN_LEGAL_PRIVACY_POLICY_CONTENT"},
				Category: configuration,
			},
			// => UserMode
			&cli.StringFlag{
				Name:     "mode",
				Usage:    "The mode of your CTFd, either users or teams.",
				Value:    "users",
				EnvVars:  []string{"MODE", "PLUGIN_MODE"},
				Category: configuration,
			},
			// => admin
			&cli.StringFlag{
				Name:     "admin.name",
				Usage:    "The administrator name. Immutable, or need the administrator to change the CTFd data AND the configuration (CLI, varenv, file). Required.",
				EnvVars:  []string{"ADMIN_NAME", "PLUGIN_ADMIN_NAME"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "admin.email",
				Usage:    "The administrator email address. Immutable, or need the administrator to change the CTFd data AND the configuration (CLI, varenv, file). Required.",
				EnvVars:  []string{"ADMIN_EMAIL", "PLUGIN_ADMIN_EMAIL"},
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "admin.password",
				Usage:    "The administrator password, recommended to use the varenvs. Immutable, or need the administrator to change the CTFd data AND the configuration (CLI, varenv, file). Required.",
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
		ctfdsetup.Log().Error("fatal error", zap.Error(err))
		os.Exit(1)
	}
}

func run(ctx *cli.Context) error {
	log := ctfdsetup.Log()

	logo, err := filePtr(ctx, "theme.logo")
	if err != nil {
		return err
	}
	smallIcon, err := filePtr(ctx, "theme.small_icon")
	if err != nil {
		return err
	}
	header, err := filePtr(ctx, "theme.header")
	if err != nil {
		return err
	}
	footer, err := filePtr(ctx, "theme.footer")
	if err != nil {
		return err
	}
	settings, err := filePtr(ctx, "theme.settings")
	if err != nil {
		return err
	}
	robotsTxt, err := filePtr(ctx, "pages.robots_txt")
	if err != nil {
		return err
	}
	conf := &ctfdsetup.Config{
		Appearance: ctfdsetup.Appearance{
			Name:          ctx.String("appearance.name"),
			Description:   ctx.String("appearance.description"),
			DefaultLocale: stringPtr(ctx, "appearance.default_locale"),
		},
		Theme: ctfdsetup.Theme{
			Logo:      logo,
			SmallIcon: smallIcon,
			Name:      ctx.String("theme.name"),
			Color:     ctx.String("theme.color"),
			Header:    header,
			Footer:    footer,
			Settings:  settings,
		},
		Accounts: ctfdsetup.Accounts{
			DomainWhitelist:               stringPtr(ctx, "accounts.domain_whitelist"),
			VerifyEmails:                  ctx.Bool("accounts.verify_emails"),
			TeamCreation:                  boolPtr(ctx, "accounts.team_creation"),
			TeamSize:                      intPtr(ctx, "accounts.team_size"),
			NumTeams:                      intPtr(ctx, "accounts.num_teams"),
			NumUsers:                      intPtr(ctx, "accounts.num_users"),
			TeamDisbanding:                stringPtr(ctx, "accounts.team_disbanding"),
			IncorrectSubmissionsPerMinute: intPtr(ctx, "accounts.incorrect_submissions_per_minute"),
			NameChanges:                   boolPtr(ctx, "accounts.name_changes"),
		},
		Pages: ctfdsetup.Pages{
			RobotsTxt: robotsTxt,
		},
		MajorLeagueCyber: ctfdsetup.MajorLeagueCyber{
			ClientID:     stringPtr(ctx, "major_league_cyber.client_id"),
			ClientSecret: stringPtr(ctx, "major_league_cyber.client_secret"),
		},
		Settings: ctfdsetup.Settings{
			ChallengeVisibility:    ctx.String("settings.challenge_visibility"),
			AccountVisibility:      ctx.String("settings.account_visibility"),
			ScoreVisibility:        ctx.String("settings.score_visibility"),
			RegistrationVisibility: ctx.String("settings.registration_visibility"),
			Paused:                 boolPtr(ctx, "settings.paused"),
		},
		Security: ctfdsetup.Security{
			HTMLSanitization: boolPtr(ctx, "security.html_sanitization"),
			RegistrationCode: stringPtr(ctx, "security.registration_code"),
		},
		Email: ctfdsetup.Email{
			Registration: ctfdsetup.EmailContent{
				Subject: stringPtr(ctx, "email.registration.subject"),
				Body:    stringPtr(ctx, "email.registration.body"),
			},
			Confirmation: ctfdsetup.EmailContent{
				Subject: stringPtr(ctx, "email.confirmation.subject"),
				Body:    stringPtr(ctx, "email.confirmation.body"),
			},
			NewAccount: ctfdsetup.EmailContent{
				Subject: stringPtr(ctx, "email.new_account.subject"),
				Body:    stringPtr(ctx, "email.new_account.body"),
			},
			PasswordReset: ctfdsetup.EmailContent{
				Subject: stringPtr(ctx, "email.password_reset.subject"),
				Body:    stringPtr(ctx, "email.password_reset.body"),
			},
			PasswordResetConfirmation: ctfdsetup.EmailContent{
				Subject: stringPtr(ctx, "email.password_reset_confirmation.subject"),
				Body:    stringPtr(ctx, "email.password_reset_confirmation.body"),
			},
			From:     stringPtr(ctx, "email.mail_from"),
			Server:   stringPtr(ctx, "email.mail_server"),
			Port:     stringPtr(ctx, "email.mail_server_port"),
			Username: stringPtr(ctx, "email.username"),
			Password: stringPtr(ctx, "email.password"),
		},
		Time: ctfdsetup.Time{
			Start:     stringPtr(ctx, "time.start"),
			End:       stringPtr(ctx, "time.end"),
			Freeze:    stringPtr(ctx, "time.freeze"),
			ViewAfter: boolPtr(ctx, "time.view_after"),
		},
		Social: ctfdsetup.Social{
			Shares: boolPtr(ctx, "social.shares"),
		},
		Legal: ctfdsetup.Legal{
			TOS: ctfdsetup.ExternalReference{
				URL:     stringPtr(ctx, "legal.tos.url"),
				Content: stringPtr(ctx, "legal.tos.content"),
			},
			PrivacyPolicy: ctfdsetup.ExternalReference{
				URL:     stringPtr(ctx, "legal.privacy_policy.url"),
				Content: stringPtr(ctx, "legal.privacy_policy.content"),
			},
		},
		Mode: ctx.String("mode"),
		Admin: ctfdsetup.Admin{
			Name:     ctx.String("admin.name"),
			Email:    ctx.String("admin.email"),
			Password: ctx.String("admin.password"),
		},
	}

	// Read and unmarshal setup config file if any
	if f := ctx.String("file"); f != "" {
		log.Info("loading configuration file", zap.String("file", f))

		fd, err := os.Open(f)
		if err != nil {
			return errors.Wrapf(err, "opening configuration file %s", f)
		}
		defer fd.Close()

		dec := yaml.NewDecoder(fd)
		dec.KnownFields(true)
		if err := dec.Decode(&conf); err != nil {
			return errors.Wrap(err, "unmarshalling configuration")
		}
	}

	if err := conf.Validate(); err != nil {
		return err
	}

	// Connect to CTFd
	return ctfdsetup.Setup(ctx.Context, ctx.String("url"), ctx.String("api_key"), conf)
}

func stringPtr(ctx *cli.Context, key string) *string {
	return genPtr(ctx, key, ctx.String)
}

func boolPtr(ctx *cli.Context, key string) *bool {
	return genPtr(ctx, key, ctx.Bool)
}

func intPtr(ctx *cli.Context, key string) *int {
	return genPtr(ctx, key, ctx.Int)
}

func filePtr(ctx *cli.Context, key string) (*ctfdsetup.File, error) {
	fp := ctx.String(key)
	if fp == "" {
		return &ctfdsetup.File{}, nil
	}
	content, err := os.ReadFile(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %s", fp)
	}
	return &ctfdsetup.File{
		Name:    filepath.Base(fp),
		Content: []byte(content),
	}, nil
}

func genPtr[T string | int | bool](ctx *cli.Context, key string, f func(key string) T) *T {
	if ctx.IsSet(key) {
		return nil
	}
	v := f(key)
	return &v
}
