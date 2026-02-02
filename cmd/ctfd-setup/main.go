package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	ctfdsetup "github.com/ctfer-io/ctfd-setup"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

const (
	management    = "management"
	configuration = "configuration"
)

func main() {
	app := &cli.Command{
		Name:  "CTFd Setup",
		Usage: "Setup (and update) a CTFd instance from a fresh install or an already-existing one.",
		Flags: []cli.Flag{
			cli.VersionFlag,
			cli.HelpFlag,
			&cli.StringFlag{
				Name:     "file",
				Usage:    "Configuration file to use for setting up CTFd. If let empty, will default the values and look for secrets in expected environment variables. For more info, refers to the documentation.",
				Sources:  cli.EnvVars("FILE", "PLUGIN_FILE"),
				Category: management,
			},
			&cli.StringFlag{
				Name:        "dir",
				Usage:       "The directory to parse from.",
				Sources:     cli.EnvVars("DIRECTORY"),
				Category:    management,
				Destination: &ctfdsetup.Directory,
			},
			&cli.StringFlag{
				Name:     "url",
				Usage:    "URL to reach the CTFd instance.",
				Sources:  cli.EnvVars("URL", "PLUGIN_URL"),
				Category: management,
			},
			&cli.StringFlag{
				Name:     "api_key",
				Usage:    "The API key to use (for instance for a CI SA), used for updating a running CTFd instance.",
				Sources:  cli.EnvVars("API_KEY", "PLUGIN_API_KEY"),
				Category: management,
			},
			// Configuration file
			// => Appearance
			&cli.StringFlag{
				Name:     "appearance.name",
				Usage:    "The name of your CTF, displayed as is.",
				Sources:  cli.EnvVars("APPEARANCE_NAME", "PLUGIN_APPEARANCE_NAME"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "appearance.description",
				Usage:    "The description of your CTF, displayed as is.",
				Sources:  cli.EnvVars("APPEARANCE_DESCRIPTION", "PLUGIN_APPEARANCE_DESCRIPTION"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "appearance.default_locale",
				Usage:    "The default language for the users.",
				Sources:  cli.EnvVars("APPEARANCE_DEFAULT_LOCALE", "PLUGIN_APPEARANCE_DEFAULT_LOCALE"),
				Category: configuration,
			},
			// => Theme
			&cli.StringFlag{
				Name:     "theme.logo",
				Usage:    "The frontend logo. Provide a path to a locally-accessible file.",
				Sources:  cli.EnvVars("THEME_LOGO", "PLUGIN_THEME_LOGO"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.small_icon",
				Usage:    "The frontend small icon. Provide a path to a locally-accessible file.",
				Sources:  cli.EnvVars("THEME_SMALL_ICON", "PLUGIN_THEME_SMALL_ICON"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.name",
				Usage:    "The frontend theme name.",
				Value:    "core",
				Sources:  cli.EnvVars("THEME_NAME", "PLUGIN_THEME_NAME"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.color",
				Usage:    "The frontend theme color.",
				Sources:  cli.EnvVars("THEME_COLOR", "PLUGIN_THEME_COLOR"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.header",
				Usage:    "The frontend header. Provide a path to a locally-accessible file.",
				Sources:  cli.EnvVars("THEME_HEADER", "PLUGIN_THEME_HEADER"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.footer",
				Usage:    "The frontend footer. Provide a path to a locally-accessible file.",
				Sources:  cli.EnvVars("THEME_FOOTER", "PLUGIN_THEME_FOOTER"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "theme.settings",
				Usage:    "The frontend settings (JSON). Provide a path to a locally-accessible file.",
				Sources:  cli.EnvVars("THEME_SETTINGS", "PLUGIN_THEME_SETTINGS"),
				Category: configuration,
			},
			// => Accounts
			&cli.StringFlag{
				Name:     "accounts.domain_whitelist",
				Usage:    "The domain whitelist (a list separated by colons) to allow users to have email addresses from.",
				Sources:  cli.EnvVars("ACCOUNTS_DOMAIN_WHITELIST", "PLUGIN_ACCOUNTS_DOMAIN_WHITELIST"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "accounts.domain_blacklist",
				Usage:    "The domain blacklist (a list separated by colons) to block users to have email addresses from.",
				Sources:  cli.EnvVars("ACCOUNTS_DOMAIN_BLACKLIST", "PLUGIN_ACCOUNTS_DOMAIN_BLACKLIST"),
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "accounts.verify_emails",
				Usage:    "Whether to verify emails once a user register or not.",
				Value:    false,
				Sources:  cli.EnvVars("ACCOUNTS_VERIFY_EMAILS", "PLUGIN_ACCOUNTS_VERIFY_EMAILS"),
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "accounts.team_creation",
				Usage:    "Whether to allow team creation by players or not.",
				Sources:  cli.EnvVars("ACCOUNTS_TEAM_CREATION", "PLUGIN_ACCOUNTS_TEAM_CREATION"),
				Category: configuration,
			},
			&cli.IntFlag{
				Name:     "accounts.team_size",
				Usage:    "Maximum size (number of players) in a team.",
				Sources:  cli.EnvVars("ACCOUNTS_TEAM_SIZE", "PLUGIN_ACCOUNTS_TEAM_SIZE"),
				Category: configuration,
			},
			&cli.IntFlag{
				Name:     "accounts.password_min_length",
				Usage:    "Minimal length of password.",
				Sources:  cli.EnvVars("ACCOUNTS_PASSWORD_MIN_LENGTH", "PLUGIN_ACCOUNTS_PASSWORD_MIN_LENGTH"),
				Category: configuration,
			},
			&cli.IntFlag{
				Name:     "accounts.num_teams",
				Usage:    "The total number of teams allowed.",
				Sources:  cli.EnvVars("ACCOUNTS_NUM_TEAMS", "PLUGIN_ACCOUNTS_NUM_TEAMS"),
				Category: configuration,
			},
			&cli.IntFlag{
				Name:     "accounts.num_users",
				Usage:    "The total number of users allowed.",
				Sources:  cli.EnvVars("ACCOUNTS_NUM_USERS", "PLUGIN_ACCOUNTS_NUM_USERS"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "accounts.team_disbanding",
				Usage:    "Whether to allow teams to be disbanded or not. Could be inactive_only or disabled.",
				Sources:  cli.EnvVars("ACCOUNTS_TEAM_DISBANDING", "PLUGIN_ACCOUNTS_TEAM_DISBANDING"),
				Category: configuration,
			},
			&cli.IntFlag{
				Name:     "accounts.incorrect_submissions_per_minute",
				Usage:    "Maximum number of invalid submissions per minute (per user/team). We suggest you use it as part of an anti-brute-force strategy (rate limiting).",
				Sources:  cli.EnvVars("ACCOUNTS_INCORRECT_SUBMISSIONS_PER_MINUTE", "PLUGIN_ACCOUNTS_INCORRECT_SUBMISSIONS_PER_MINUTE"),
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "accounts.name_changes",
				Usage:    "Whether a user can change its name or not.",
				Sources:  cli.EnvVars("ACCOUNTS_NAME_CHANGES", "PLUGIN_ACCOUNTS_NAME_CHANGES"),
				Category: configuration,
			},
			// => Challenges
			&cli.BoolFlag{
				Name:     "challenges.view_self_submissions",
				Usage:    "Whether a player can see itw own previous submissions.",
				Sources:  cli.EnvVars("CHALLENGES_VIEW_SELF_SUBMISSIONS", "PLUGIN_CHALLENGES_VIEW_SELF_SUBMISSIONS"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "challenges.max_attempts_behavior",
				Usage:    "The behavior to adopt in case a player reached the submission rate limiting.",
				Value:    "lockout",
				Sources:  cli.EnvVars("CHALLENGES_MAX_ATTEMPTS_BEHAVIOR", "PLUGIN_CHALLENGES_MAX_ATTEMPTS_BEHAVIOR"),
				Category: configuration,
			},
			&cli.IntFlag{
				Name:     "challenges.max_attempts_timeout",
				Usage:    "The duration of the submission rate limit for further submissions.",
				Sources:  cli.EnvVars("CHALLENGES_MAX_ATTEMPTS_TIMEOUT", "PLUGIN_CHALLENGES_MAX_ATTEMPTS_TIMEOUT"),
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "challenges.hints_free_public_access",
				Usage:    "Control whether users must be logged in to see free hints.",
				Sources:  cli.EnvVars("CHALLENGES_HINTS_FREE_PUBLIC_ACCESS", "PUBLIC_CHALLENGES_HINTS_FREE_PUBLIC_ACCESS"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "challenges.challenge_ratings",
				Usage:    "Who can see and submit challenge ratings.",
				Value:    "public",
				Sources:  cli.EnvVars("CHALLENGES_CHALLENGE_RATINGS", "PUBLIC_CHALLENGES_CHALLENGE_RATINGS"),
				Category: configuration,
			},
			// => Pages
			&cli.StringFlag{
				Name:     "pages.robots_txt",
				Usage:    "Define the /robots.txt file content, for web crawlers indexing. Provide a path to a locally-accessible file.",
				Sources:  cli.EnvVars("PAGES_ROBOTS_TXT", "PLUGIN_PAGES_ROBOTS_TXT"),
				Category: configuration,
			},
			// => MajorLeagueCyber
			&cli.StringFlag{
				Name:     "major_league_cyber.client_id",
				Usage:    "The MajorLeagueCyber OAuth ClientID.",
				Sources:  cli.EnvVars("MAJOR_LEAGUE_CYBER_CLIENT_ID", "PLUGIN_MAJOR_LEAGUE_CYBER_CLIENT_ID"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "major_league_cyber.client_secret",
				Usage:    "The MajorLeagueCyber OAuth Client Secret.",
				Sources:  cli.EnvVars("MAJOR_LEAGUE_CYBER_CLIENT_SECRET", "PLUGIN_MAJOR_LEAGUE_CYBER_CLIENT_SECRET"),
				Category: configuration,
			},
			// => Settings
			&cli.StringFlag{
				Name:     "settings.challenge_visibility",
				Usage:    "The visibility for the challenges. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).",
				Value:    "public",
				Sources:  cli.EnvVars("SETTINGS_CHALLENGE_VISIBILITY", "PLUGIN_SETTINGS_CHALLENGE_VISIBILITY"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "settings.account_visibility",
				Usage:    "The visibility for the accounts. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).",
				Value:    "public",
				Sources:  cli.EnvVars("SETTINGS_ACCOUNT_VISIBILITY", "PLUGIN_SETTINGS_ACCOUNT_VISIBILITY"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "settings.score_visibility",
				Usage:    "The visibility for the scoreboard. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).",
				Value:    "public",
				Sources:  cli.EnvVars("SETTINGS_SCORE_VISIBILITY", "PLUGIN_SETTINGS_SCORE_VISIBILITY"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "settings.registration_visibility",
				Usage:    "The visibility for the registration. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).",
				Value:    "public",
				Sources:  cli.EnvVars("SETTINGS_REGISTRATION_VISIBILITY", "PLUGIN_SETTINGS_REGISTRATION_VISIBILITY"),
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "settings.paused",
				Usage:    "Whether the CTFd is paused or not.",
				Sources:  cli.EnvVars("SETTINGS_PAUSED", "PLUGIN_SETTINGS_PAUSED"),
				Category: configuration,
			},
			// => Security
			&cli.BoolFlag{
				Name:     "security.html_sanitization",
				Usage:    "Whether to turn on HTML sanitization or not.",
				Sources:  cli.EnvVars("SECURITY_HTML_SANITIZATION", "PLUGIN_SECURITY_HTML_SANITIZATION"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "security.registration_code",
				Usage:    "The registration code (secret) to join the CTF.",
				Sources:  cli.EnvVars("SECURITY_REGISTRATION_CODE", "PLUGIN_SECURITY_REGISTRATION_CODE"),
				Category: configuration,
			},
			// => Email
			&cli.StringFlag{
				Name:     "email.registration.subject",
				Usage:    "The email registration subject of the mail.",
				Sources:  cli.EnvVars("EMAIL_REGISTRATION_SUBJECT", "PLUGIN_EMAIL_REGISTRATION_SUBJECT"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.registration.body",
				Usage:    "The email registration body of the mail.",
				Sources:  cli.EnvVars("EMAIL_REGISTRATION_BODY", "PLUGIN_EMAIL_REGISTRATION_BODY"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.confirmation.subject",
				Usage:    "The email confirmation subject of the mail.",
				Sources:  cli.EnvVars("EMAIL_CONFIRMATION_SUBJECT", "PLUGIN_EMAIL_CONFIRMATION_SUBJECT"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.confirmation.body",
				Usage:    "The email confirmation body of the mail.",
				Sources:  cli.EnvVars("EMAIL_CONFIRMATION_BODY", "PLUGIN_EMAIL_CONFIRMATION_BODY"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.new_account.subject",
				Usage:    "The email new_account subject of the mail.",
				Sources:  cli.EnvVars("EMAIL_NEW_ACCOUNT_SUBJECT", "PLUGIN_EMAIL_NEW_ACCOUNT_SUBJECT"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.new_account.body",
				Usage:    "The email new_account body of the mail.",
				Sources:  cli.EnvVars("EMAIL_NEW_ACCOUNT_BODY", "PLUGIN_EMAIL_NEW_ACCOUNT_BODY"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.password_reset.subject",
				Usage:    "The email password_reset subject of the mail.",
				Sources:  cli.EnvVars("EMAIL_PASSWORD_RESET_SUBJECT", "PLUGIN_EMAIL_PASSWORD_RESET_SUBJECT"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.password_reset.body",
				Usage:    "The email password_reset body of the mail.",
				Sources:  cli.EnvVars("EMAIL_PASSWORD_RESET_BODY", "PLUGIN_EMAIL_PASSWORD_RESET_BODY"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.password_reset_confirmation.subject",
				Usage:    "The email password_reset_confirmation subject of the mail.",
				Sources:  cli.EnvVars("EMAIL_PASSWORD_RESET_CONFIRMATION_SUBJECT", "PLUGIN_EMAIL_PASSWORD_RESET_CONFIRMATION_SUBJECT"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.password_reset_confirmation.body",
				Usage:    "The email password_reset_confirmation body of the mail.",
				Sources:  cli.EnvVars("EMAIL_PASSWORD_RESET_CONFIRMATION_BODY", "PLUGIN_EMAIL_PASSWORD_RESET_CONFIRMATION_BODY"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.from",
				Usage:    "The 'From:' to sent to mail with.",
				Sources:  cli.EnvVars("EMAIL_MAIL_FROM", "PLUGIN_EMAIL_MAIL_FROM"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.server",
				Usage:    "The mail server to use.",
				Sources:  cli.EnvVars("EMAIL_MAIL_SERVER", "PLUGIN_EMAIL_MAIL_SERVER"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.port",
				Usage:    "The mail server port to reach.",
				Sources:  cli.EnvVars("EMAIL_MAIL_SERVER_PORT", "PLUGIN_EMAIL_MAIL_SERVER_PORT"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.username",
				Usage:    "The username to log in to the mail server.",
				Sources:  cli.EnvVars("EMAIL_USERNAME", "PLUGIN_EMAIL_USERNAME"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "email.password",
				Usage:    "The password to log in to the mail server.",
				Sources:  cli.EnvVars("EMAIL_PASSWORD", "PLUGIN_EMAIL_PASSWORD"),
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "email.tls_ssl",
				Usage:    "Whether to turn on TLS/SSL or not.",
				Sources:  cli.EnvVars("EMAIL_TLS_SSL", "PLUGIN_EMAIL_TLS_SSL"),
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "email.starttls",
				Usage:    "Whether to turn on STARTTLS or not.",
				Sources:  cli.EnvVars("EMAIL_STARTTLS", "PLUGIN_EMAIL_STARTTLS"),
				Category: configuration,
			},
			// => Time
			&cli.StringFlag{
				Name:     "time.start",
				Usage:    "The start timestamp at which the CTFd will open.",
				Sources:  cli.EnvVars("TIME_START", "PLUGIN_TIME_START"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "time.end",
				Usage:    "The end timestamp at which the CTFd will close.",
				Sources:  cli.EnvVars("TIME_END", "PLUGIN_TIME_END"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "time.freeze",
				Usage:    "The freeze timestamp at which the CTFd will remain open but won't accept any further submissions.",
				Sources:  cli.EnvVars("TIME_FREEZE", "PLUGIN_TIME_FREEZE"),
				Category: configuration,
			},
			&cli.BoolFlag{
				Name:     "time.view_after",
				Usage:    "Whether allows users to view challenges after end or not.",
				Sources:  cli.EnvVars("TIME_VIEW_AFTER", "PLUGIN_TIME_VIEW_AFTER"),
				Category: configuration,
			},
			// => Social
			&cli.BoolFlag{
				Name:     "social.shares",
				Usage:    "Whether to enable users share they solved a challenge or not.",
				Sources:  cli.EnvVars("SOCIAL_SHARES", "PLUGIN_SOCIAL_SHARES"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "social.template",
				Usage:    "A template for social shares. Provide a path to a locally-accessible file.",
				Sources:  cli.EnvVars("SOCIAL_TEMPLATE", "PUBLIC_SOCIAL_TEMPLATE"),
				Category: configuration,
			},
			// => Legal
			&cli.StringFlag{
				Name:     "legal.tos.url",
				Usage:    "The Terms of Services URL.",
				Sources:  cli.EnvVars("LEGAL_TOS_URL", "PLUGIN_LEGAL_TOS_URL"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "legal.tos.content",
				Usage:    "The Terms of Services content.",
				Sources:  cli.EnvVars("LEGAL_TOS_CONTENT", "PLUGIN_LEGAL_TOS_CONTENT"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "legal.privacy_policy.url",
				Usage:    "The Privacy Policy URL.",
				Sources:  cli.EnvVars("LEGAL_PRIVACY_POLICY_URL", "PLUGIN_LEGAL_PRIVACY_POLICY_URL"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "legal.privacy_policy.content",
				Usage:    "The Privacy Policy content.",
				Sources:  cli.EnvVars("LEGAL_PRIVACY_POLICY_CONTENT", "PLUGIN_LEGAL_PRIVACY_POLICY_CONTENT"),
				Category: configuration,
			},
			// => UserMode
			&cli.StringFlag{
				Name:     "mode",
				Usage:    "The mode of your CTFd, either users or teams.",
				Value:    "users",
				Sources:  cli.EnvVars("MODE", "PLUGIN_MODE"),
				Category: configuration,
			},
			// => admin
			&cli.StringFlag{
				Name:     "admin.name",
				Usage:    "The administrator name. Immutable, or need the administrator to change the CTFd data AND the configuration (CLI, varenv, file). Required.",
				Sources:  cli.EnvVars("ADMIN_NAME", "PLUGIN_ADMIN_NAME"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "admin.email",
				Usage:    "The administrator email address. Immutable, or need the administrator to change the CTFd data AND the configuration (CLI, varenv, file). Required.",
				Sources:  cli.EnvVars("ADMIN_EMAIL", "PLUGIN_ADMIN_EMAIL"),
				Category: configuration,
			},
			&cli.StringFlag{
				Name:     "admin.password",
				Usage:    "The administrator password, recommended to use the varenvs. Immutable, or need the administrator to change the CTFd data AND the configuration (CLI, varenv, file). Required.",
				Sources:  cli.EnvVars("ADMIN_PASSWORD", "PLUGIN_ADMIN_PASSWORD"),
				Category: configuration,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "schema",
				Usage: "Generate the JSON schema of a .ctfd.yaml file.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "The output file name.",
						Value:   "schema.json",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					o := cmd.String("output")
					schema, err := ctfdsetup.Config{}.Schema()
					if err != nil {
						return err
					}
					return os.WriteFile(o, schema, os.ModeAppend|os.ModePerm)
				},
			},
		},
		Action: run,
		Authors: []any{
			"CTFer.io's authors and contributors - ctfer-io@protonmail.com",
		},
		Version: version,
		Metadata: map[string]any{
			"version": version,
			"commit":  commit,
			"date":    date,
			"builtBy": builtBy,
		},
	}

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, os.Args); err != nil {
		ctfdsetup.Log().Error("fatal error", zap.Error(err))
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	log := ctfdsetup.Log()

	shutdown, err := ctfdsetup.SetupOtelSDK(ctx, version)
	if err != nil {
		return err
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Error("shuttding down tracer provider",
				zap.Error(err),
			)
		}
	}()

	logo, err := filePtr(cmd, "theme.logo")
	if err != nil {
		return err
	}
	smallIcon, err := filePtr(cmd, "theme.small_icon")
	if err != nil {
		return err
	}
	header, err := filePtr(cmd, "theme.header")
	if err != nil {
		return err
	}
	footer, err := filePtr(cmd, "theme.footer")
	if err != nil {
		return err
	}
	settings, err := filePtr(cmd, "theme.settings")
	if err != nil {
		return err
	}
	robotsTxt, err := filePtr(cmd, "pages.robots_txt")
	if err != nil {
		return err
	}
	socialTpl, err := filePtr(cmd, "social.template")
	if err != nil {
		return err
	}

	tos, err := filePtr(cmd, "legal.tos.content")
	if err != nil {
		return err
	}
	privpol, err := filePtr(cmd, "legal.privacy_policy.content")
	if err != nil {
		return err
	}
	conf := ctfdsetup.NewConfig()

	// Read and unmarshal setup config file if any
	if f := cmd.String("file"); f != "" {
		log.Info("loading configuration file", zap.String("file", f))

		fd, err := os.Open(f)
		if err != nil {
			return errors.Wrapf(err, "opening configuration file %s", f)
		}
		defer func() {
			_ = fd.Close()
		}()

		dec := yaml.NewDecoder(fd)
		dec.KnownFields(true)
		if err := dec.Decode(&conf); err != nil {
			return errors.Wrap(err, "unmarshalling configuration")
		}
	}

	// Then override with all CLI flags.
	// This is especially usefull to define placeholders in configuration file but
	// use real credentials provided from environment variables within CI (or any other automation).
	overrideForDefaultString(cmd, &conf.Appearance.Name, "appearance.name")
	overrideForDefaultString(cmd, &conf.Appearance.Description, "appearance.description")
	overrideForDefaultStringPtr(cmd, &conf.Appearance.DefaultLocale, "appearance.default_locale")

	conf.Theme.Logo = logo
	conf.Theme.SmallIcon = smallIcon
	overrideForDefaultString(cmd, &conf.Theme.Name, "theme.name")
	overrideForDefaultString(cmd, &conf.Theme.Color, "theme.color")
	conf.Theme.Header = header
	conf.Theme.Footer = footer
	conf.Theme.Settings = settings

	overrideForDefaultStringPtr(cmd, &conf.Accounts.DomainWhitelist, "accounts.domain_whitelist")
	overrideForDefaultStringPtr(cmd, &conf.Accounts.DomainBlacklist, "accounts.domain_blacklist")
	overrideForDefaultBool(cmd, &conf.Accounts.VerifyEmails, "accounts.verify_emails")
	overrideForDefaultBoolPtr(cmd, &conf.Accounts.TeamCreation, "accounts.team_creation")
	overrideForDefaultIntPtr(cmd, &conf.Accounts.TeamSize, "accounts.team_size")
	overrideForDefaultIntPtr(cmd, &conf.Accounts.PasswordMinLength, "accounts.password_min_length")
	overrideForDefaultIntPtr(cmd, &conf.Accounts.NumTeams, "accounts.num_teams")
	overrideForDefaultIntPtr(cmd, &conf.Accounts.NumUsers, "accounts.num_users")
	overrideForDefaultStringPtr(cmd, &conf.Accounts.TeamDisbanding, "accounts.team_disbanding")
	overrideForDefaultIntPtr(cmd, &conf.Accounts.IncorrectSubmissionsPerMinute, "accounts.incorrect_submissions_per_minute")
	overrideForDefaultBoolPtr(cmd, &conf.Accounts.NameChanges, "accounts.name_changes")

	overrideForDefaultBool(cmd, &conf.Challenges.ViewSelfSubmission, "challenges.view_self_submissions")
	overrideForDefaultString(cmd, &conf.Challenges.MaxAttemptsBehavior, "challenges.max_attempts_behavior")
	overrideForDefaultInt(cmd, &conf.Challenges.MaxAttemptsTimeout, "challenges.max_attempts_timeout")
	overrideForDefaultBool(cmd, &conf.Challenges.HintsFreePublicAccess, "challenges.hints_free_public_access")
	overrideForDefaultString(cmd, &conf.Challenges.ChallengeRatings, "challenges.challenge_ratings")

	conf.Pages.RobotsTxt = robotsTxt

	overrideForDefaultStringPtr(cmd, &conf.MajorLeagueCyber.ClientID, "major_league_cyber.client_id")
	overrideForDefaultStringPtr(cmd, &conf.MajorLeagueCyber.ClientSecret, "major_league_cyber.client_secret")

	overrideForDefaultString(cmd, &conf.Settings.ChallengeVisibility, "settings.challenge_visibility")
	overrideForDefaultString(cmd, &conf.Settings.AccountVisibility, "settings.account_visibility")
	overrideForDefaultString(cmd, &conf.Settings.ScoreVisibility, "settings.score_visibility")
	overrideForDefaultString(cmd, &conf.Settings.RegistrationVisibility, "settings.registration_visibility")
	overrideForDefaultBoolPtr(cmd, &conf.Settings.Paused, "settings.paused")

	overrideForDefaultBoolPtr(cmd, &conf.Security.HTMLSanitization, "security.html_sanitization")
	overrideForDefaultStringPtr(cmd, &conf.Security.RegistrationCode, "security.registration_code")

	overrideForDefaultStringPtr(cmd, &conf.Email.Registration.Subject, "email.registration.subject")
	overrideForDefaultStringPtr(cmd, &conf.Email.Registration.Body, "email.registration.body")
	overrideForDefaultStringPtr(cmd, &conf.Email.Confirmation.Subject, "email.confirmation.subject")
	overrideForDefaultStringPtr(cmd, &conf.Email.Confirmation.Body, "email.confirmation.body")
	overrideForDefaultStringPtr(cmd, &conf.Email.NewAccount.Subject, "email.new_account.subject")
	overrideForDefaultStringPtr(cmd, &conf.Email.NewAccount.Body, "email.new_account.body")
	overrideForDefaultStringPtr(cmd, &conf.Email.PasswordReset.Subject, "email.password_reset.subject")
	overrideForDefaultStringPtr(cmd, &conf.Email.PasswordReset.Body, "email.password_reset.body")
	overrideForDefaultStringPtr(cmd, &conf.Email.PasswordResetConfirmation.Subject, "email.password_reset_confirmation.subject")
	overrideForDefaultStringPtr(cmd, &conf.Email.PasswordResetConfirmation.Body, "email.password_reset_confirmation.body")
	overrideForDefaultStringPtr(cmd, &conf.Email.From, "email.mail_from")
	overrideForDefaultStringPtr(cmd, &conf.Email.Server, "email.mail_server")
	overrideForDefaultStringPtr(cmd, &conf.Email.Port, "email.mail_server_port")
	overrideForDefaultStringPtr(cmd, &conf.Email.Username, "email.username")
	overrideForDefaultStringPtr(cmd, &conf.Email.Password, "email.password")

	overrideForDefaultStringPtr(cmd, &conf.Time.Start, "time.start")
	overrideForDefaultStringPtr(cmd, &conf.Time.End, "time.end")
	overrideForDefaultStringPtr(cmd, &conf.Time.Freeze, "time.freeze")
	overrideForDefaultBoolPtr(cmd, &conf.Time.ViewAfter, "time.view_after")

	overrideForDefaultBoolPtr(cmd, &conf.Social.Shares, "social.shares")
	conf.Social.Template = socialTpl

	overrideForDefaultStringPtr(cmd, &conf.Legal.TOS.URL, "legal.tos.url")
	conf.Legal.TOS.Content = tos
	overrideForDefaultStringPtr(cmd, &conf.Legal.PrivacyPolicy.URL, "legal.privacy_policy.url")
	conf.Legal.PrivacyPolicy.Content = privpol

	overrideForDefaultString(cmd, &conf.Mode, "mode")

	overrideForDefaultString(cmd, &conf.Admin.Name.Content, "admin.name")
	overrideForDefaultString(cmd, &conf.Admin.Email.Content, "admin.email")
	overrideForDefaultString(cmd, &conf.Admin.Password.Content, "admin.password")

	if err := conf.Validate(); err != nil {
		return err
	}

	// Connect to CTFd
	if !cmd.IsSet("url") {
		return errors.New("url flag not set, is required")
	}
	return ctfdsetup.Setup(ctx, cmd.String("url"), cmd.String("api_key"), conf)
}

func filePtr(cmd *cli.Command, key string) (*ctfdsetup.File, error) {
	fp := cmd.String(key)
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

func overrideForDefaultString(cmd *cli.Command, dst *string, key string) {
	str := cmd.String(key)
	if cmd.IsSet(key) && str != "" { // avoid empty strings
		*dst = str
	}
	// Don't change anything, it remains as it is
}

func overrideForDefaultStringPtr(cmd *cli.Command, dst **string, key string) {
	str := cmd.String(key)
	if cmd.IsSet(key) && str != "" { // avoid empty strings
		*dst = &str
	}
	// Don't change anything, it remains as it is
}

func overrideForDefaultBool(cmd *cli.Command, dst *bool, key string) {
	b := cmd.Bool(key)
	if cmd.IsSet(key) && b != false { // avoid false values
		*dst = b
	}
	// Don't change anything, it remains as it is
}

func overrideForDefaultBoolPtr(cmd *cli.Command, dst **bool, key string) {
	b := cmd.Bool(key)
	if cmd.IsSet(key) && b != false { // avoid false values
		*dst = &b
	}
	// Don't change anything, it remains as it is
}

func overrideForDefaultInt(cmd *cli.Command, dst *int, key string) {
	i := cmd.Int(key)
	if cmd.IsSet(key) && i != 0 { // avoid zeros
		*dst = i
	}
	// Don't change anything, it remains as it is
}

func overrideForDefaultIntPtr(cmd *cli.Command, dst **int, key string) {
	i := cmd.Int(key)
	if cmd.IsSet(key) && i != 0 { // avoid zeros
		*dst = &i
	}
	// Don't change anything, it remains as it is
}
