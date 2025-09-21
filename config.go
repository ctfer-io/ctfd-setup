package ctfdsetup

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/multierr"
)

type (
	Config struct {
		// Don't handle brackets here, should not be part of those settings but CRUD objects
		// CustomFields are not handled as they are not predictable and would be hard to handle + bad practice (API changes on the fly)

		Appearance       Appearance        `yaml:"appearance"                   json:"appearance"                   jsonschema:"required"`
		Theme            *Theme            `yaml:"theme,omitempty"              json:"theme,omitempty"`
		Accounts         *Accounts         `yaml:"accounts,omitempty"           json:"accounts,omitempty"`
		Challenges       *Challenges       `yaml:"challenges,omitempty"         json:"challenges,omitempty"`
		Pages            *Pages            `yaml:"pages,omitempty"              json:"pages,omitempty"`
		MajorLeagueCyber *MajorLeagueCyber `yaml:"major_league_cyber,omitempty" json:"major_league_cyber,omitempty"`
		Settings         *Settings         `yaml:"settings,omitempty"           json:"settings,omitempty"`
		Security         *Security         `yaml:"security,omitempty"           json:"security,omitempty"`
		Email            *Email            `yaml:"email,omitempty"              json:"email,omitempty"`
		Time             *Time             `yaml:"time,omitempty"               json:"time,omitempty"`
		Social           *Social           `yaml:"social,omitempty"             json:"social,omitempty"`
		Legal            *Legal            `yaml:"legal,omitempty"              json:"legal,omitempty"`
		Admin            Admin             `yaml:"admin"                        json:"admin"                        jsonschema:"required"`

		// The mode of your CTFd, either users or teams.
		Mode string `yaml:"mode,omitempty" json:"mode,omitempty" jsonschema:"enum=users,enum=teams,default=users"`

		Uploads []*Upload `yaml:"uploads,omitempty" json:"uploads,omitempty"`
	}

	// Appearance of the CTFd.
	Appearance struct {
		// The name of your CTF, displayed as is.
		Name string `yaml:"name" json:"name" jsonschema:"required"`

		// The description of your CTF, displayed as is.
		Description string `yaml:"description" json:"description" jsonschema:"required"`

		// The default language for the users.
		DefaultLocale *string `yaml:"default_locale,omitempty" json:"default_locale,omitempty"`
	}

	// Theme displayed to end-users.
	Theme struct {
		// Banner is only supported by bare setup, need to be at least support by PatchConfigs

		// The frontend logo.
		Logo *File `yaml:"logo,omitempty" json:"logo,omitempty"`

		// The frontend small icon.
		SmallIcon *File `yaml:"small_icon,omitempty" json:"small_icon,omitempty"`

		// The frontend theme name.
		Name string `yaml:"name,omitempty" json:"name,omitempty" jsonschema:"default=core-beta"` // do not restrict to core-beta or core (deprecated) to avoid limiting to official themes

		// The frontend theme color.
		Color string `yaml:"color,omitempty" json:"color,omitempty"`

		// The frontend header.
		Header *File `yaml:"header,omitempty" json:"header,omitempty"`

		// The frontend footer.
		Footer *File `yaml:"footer,omitempty" json:"footer,omitempty"`

		// The frontend settings (JSON).
		Settings *File `yaml:"settings,omitempty" json:"settings,omitempty"`
	}

	// Accounts parameters, like rate limiting or default permissions.
	Accounts struct {
		// The domain whitelist (a list separated by colons) to allow users to have email addresses from.
		DomainWhitelist *string `yaml:"domain_whitelist,omitempty" json:"domain_whitelist,omitempty"`

		// The domain blacklist (a list separated by colons) to blocks users to have email addresses from.
		DomainBlacklist *string `yaml:"domain_blacklist,omitempty" json:"domain_blacklist,omitempty"`

		// Whether to verify emails once a user register or not.
		VerifyEmails bool `yaml:"verify_emails,omitempty" json:"verify_emails,omitempty"`

		// Whether to allow team creation by players or not.
		TeamCreation *bool `yaml:"team_creation,omitempty" json:"team_creation,omitempty"`

		// Maximum size (number of players) in a team.
		TeamSize *int `yaml:"team_size,omitempty" json:"team_size,omitempty"`

		// Minimal length of passwords.
		PasswordMinLength *int `yaml:"password_min_length,omitempty" json:"password_min_length,omitempty"`

		// The total number of teams allowed.
		NumTeams *int `yaml:"num_teams,omitempty" json:"num_teams,omitempty"`

		// The total number of users allowed.
		NumUsers *int `yaml:"num_users,omitempty" json:"num_users,omitempty"`

		// Whether to allow teams to be disbanded or not. Could be inactive_only or disabled.
		TeamDisbanding *string `yaml:"team_disbanding,omitempty" json:"team_disbanding,omitempty"`

		// Maximum number of invalid submissions per minute (per user/team). We suggest you use it as part of an anti-brute-force strategy (rate limiting).
		IncorrectSubmissionsPerMinute *int `yaml:"incorrect_submissions_per_minute,omitempty" json:"incorrect_submissions_per_minute,omitempty"`

		// Whether a user can change its name or not.
		NameChanges *bool `yaml:"name_changes,omitempty" json:"name_changes,omitempty"`
	}

	// Challenge-related configurations.
	Challenges struct {
		// Whether a player can see itw own previous submissions.
		ViewSelfSubmission bool `yaml:"view_self_submissions" json:"view_self_submissions"`

		// The behavior to adopt in case a player reached the submission rate limiting.
		MaxAttemptsBehavior string `yaml:"max_attempts_behavior" json:"max_attempts_behavior" jsonschema:"enum=lockout,enum=timeout,default=lockout"`

		// The duration of the submission rate limit for further submissions.
		MaxAttemptsTimeout int `yaml:"max_attempts_timeout" json:"max_attempts_timeout"`

		// Control whether users must be logged in to see free hints.
		HintsFreePublicAccess bool `yaml:"hints_free_public_access" json:"hints_free_public_access"`

		// Who can see and submit challenge ratings.
		ChallengeRatings string `yaml:"challenge_ratings" json:"challenge_ratings" jsonschema:"enum=public,enum=private,enum=disabled,default=public"`
	}

	// Pages global configuration.
	Pages struct {
		// Define the /robots.txt file content, for web crawlers indexing.
		RobotsTxt *File `yaml:"robots_txt,omitempty" json:"robots_txt,omitempty"`

		Additional []Page `yaml:"additional,omitempty" json:"additional,omitempty"`
	}

	// Page to configure and display on the CTFd.
	Page struct {
		// Title of the page.
		Title string `yaml:"title" json:"title"`

		// Route to serve.
		Route string `yaml:"route" json:"route"`

		// Format to consume the content.
		Format string `yaml:"format,omitempty" json:"format,omitempty" jsonschema:"enum=markdown,enum=html,default=markdown"`

		// The page content. If you need to use images, please use an external CDN to make sure the content is replicable.
		Content *File `yaml:"content" json:"content"`

		// Set the page as a draft.
		Draft bool `yaml:"draft,omitempty" json:"draft,omitempty" jsonschema:"default=false"`

		// Hide or show the page to users.
		Hidden bool `yaml:"hidden,omitempty" json:"hidden,omitempty" jsonschema:"default=false"`

		// Configure whether the page require authentication or not.
		AuthRequired bool `yaml:"auth_required,omitempty" json:"auth_required,omitempty" jsonschema:"default=false"`
	}

	// MajorLeagueCyber credentials to register the CTF.
	MajorLeagueCyber struct {
		// The MajorLeagueCyber OAuth ClientID.
		ClientID *string `yaml:"client_id,omitempty" json:"client_id,omitempty"`

		// The MajorLeagueCyber OAuth Client Secret.
		ClientSecret *string `yaml:"client_secret,omitempty" json:"client_secret,omitempty"`
	}

	// Settings for resources visibility.
	Settings struct {
		// The visibility for the challenges. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).
		ChallengeVisibility string `yaml:"challenge_visibility,omitempty" json:"challenge_visibility,omitempty" jsonschema:"enum=public,enum=private,enum=admins,default=private"`

		// The visibility for the accounts. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).
		AccountVisibility string `yaml:"account_visibility,omitempty" json:"account_visibility,omitempty" jsonschema:"enum=public,enum=private,enum=admins,default=public"`

		// The visibility for the scoreboard. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).
		ScoreVisibility string `yaml:"score_visibility,omitempty" json:"score_visibility,omitempty" jsonschema:"enum=public,enum=private,enum=admins,default=public"`

		// The visibility for the registration. Please refer to CTFd documentation (https://docs.ctfd.io/docs/settings/visibility-settings/).
		RegistrationVisibility string `yaml:"registration_visibility,omitempty" json:"registration_visibility,omitempty" jsonschema:"enum=public,enum=private,enum=admins,default=public"`

		// Whether the CTFd is paused or not.
		Paused *bool `yaml:"paused,omitempty" json:"paused,omitempty"`
	}

	// Security of contents and accesses.
	Security struct {
		// Whether to turn on HTML sanitization or not.
		HTMLSanitization *bool `yaml:"html_sanitization,omitempty" json:"html_sanitization,omitempty"`

		// The registration code (secret) to join the CTF.
		RegistrationCode *string `yaml:"registration_code,omitempty" json:"registration_code,omitempty"`
	}

	// Email rules and server credentials.
	Email struct {
		// The registration email.
		Registration EmailContent `yaml:"registration,omitempty" json:"registration,omitempty"`

		// The confirmation email.
		Confirmation EmailContent `yaml:"confirmation,omitempty" json:"confirmation,omitempty"`

		// The new account email.
		NewAccount EmailContent `yaml:"new_account,omitempty" json:"new_account,omitempty"`

		// The password reset email.
		PasswordReset EmailContent `yaml:"password_reset,omitempty" json:"password_reset,omitempty"`

		// The password reset confirmation email.
		PasswordResetConfirmation EmailContent `yaml:"password_reset_confirmation,omitempty" json:"password_reset_confirmation,omitempty"`

		// The 'From:' to sent to mail with.
		From *string `yaml:"from,omitempty" json:"from,omitempty"`

		// The mail server to use.
		Server *string `yaml:"server,omitempty" json:"server,omitempty"`

		// The mail server port to reach.
		Port *string `yaml:"port,omitempty" json:"port,omitempty"`

		// The username to log in to the mail server.
		Username *string `yaml:"username,omitempty" json:"username,omitempty"`

		// The password to log in to the mail server.
		Password *string `yaml:"password,omitempty" json:"password,omitempty"`

		// Whether to turn on TLS/SSL or not.
		TLS_SSL *bool `yaml:"tls_ssl,omitempty" json:"tls_ssl,omitempty"`

		// Whether to turn on STARTTLS or not.
		STARTTLS *bool `yaml:"starttls,omitempty" json:"starttls,omitempty"`
	}

	EmailContent struct {
		// Subject of the email.
		Subject *string `yaml:"subject,omitempty" json:"subject,omitempty"`

		// Body (or content) or the email.
		Body *string `yaml:"body,omitempty"    json:"body,omitempty"`
	}

	// Time settings of the CTF.
	Time struct {
		// The start timestamp at which the CTFd will open.
		Start *string `yaml:"start,omitempty" json:"start,omitempty"`

		// The end timestamp at which the CTFd will close.
		End *string `yaml:"end,omitempty" json:"end,omitempty"`

		// The freeze timestamp at which the CTFd will remain open but won't accept any further submissions.
		Freeze *string `yaml:"freeze,omitempty" json:"freeze,omitempty"`

		// Whether allows users to view challenges after end or not.
		ViewAfter *bool `yaml:"view_after,omitempty" json:"view_after,omitempty"`
	}

	// Social network configuration.
	Social struct {
		// Whether to enable users share they solved a challenge or not.
		Shares *bool `yaml:"shares,omitempty" json:"shares,omitempty"`

		// A template for social shares.
		Template *File `yaml:"template,omitempty" json:"template,omitempty"`
	}

	// Legal contents for players.
	Legal struct {
		// The Terms of Services.
		TOS ExternalReference `yaml:"tos,omitempty" json:"tos,omitempty"`

		// The Privacy Policy.
		PrivacyPolicy ExternalReference `yaml:"privacy_policy,omitempty" json:"privacy_policy,omitempty"`
	}

	ExternalReference struct {
		// The URL to access the content.
		URL *string `yaml:"url,omitempty" json:"url,omitempty"`

		// The content of the reference.
		Content *File `yaml:"content,omitempty" json:"content,omitempty"`
	}

	// Admin accesses.
	Admin struct {
		// The administrator name. Immutable, or need the administrator to change the CTFd data AND the configuration file.
		Name string `yaml:"name,omitempty" json:"name,omitempty" jsonschema:"required"`

		// The administrator email address. Immutable, or need the administrator to change the CTFd data AND the configuration file.
		Email string `yaml:"email,omitempty" json:"email,omitempty" jsonschema:"required"`

		// The administrator password, recommended to use the varenvs. Immutable, or need the administrator to change the CTFd data AND the configuration file.
		Password FromEnv `yaml:"password,omitempty" json:"password,omitempty" jsonschema:"required"`
	}

	// Upload defines a file or content to upload as per the setup. Does not upload twice if already exist.
	// One use case is to upload logos and use them in an alternative index.html page for an event.
	//
	// WARNING: if a file is removed from the list, it won't be deleted by ctfd-setup.
	Upload struct {
		File *File `yaml:"file" json:"file" jsonschema:"required"`

		// Where to upload it.
		// This enables to use a file at a static location in, e.g., custom pages.
		Location string `yaml:"location" json:"location" jsonschema:"required"`
	}
)

func NewConfig() *Config {
	return &Config{
		Theme: &Theme{
			Name:      "core-beta",
			Logo:      &File{},
			SmallIcon: &File{},
			Header:    &File{},
			Footer:    &File{},
			Settings:  &File{},
		},
		Accounts: &Accounts{},
		Challenges: &Challenges{
			MaxAttemptsBehavior: "lockout",
			ChallengeRatings:    "public",
		},
		Pages: &Pages{
			RobotsTxt: &File{},
		},
		MajorLeagueCyber: &MajorLeagueCyber{},
		Settings: &Settings{
			ChallengeVisibility:    "private", // default value
			AccountVisibility:      "public",  // default value
			ScoreVisibility:        "public",  // default value
			RegistrationVisibility: "public",  // default value
		},
		Security: &Security{},
		Email:    &Email{},
		Time:     &Time{},
		Social: &Social{
			Template: &File{},
		},
		Legal: &Legal{
			TOS: ExternalReference{
				Content: &File{},
			},
			PrivacyPolicy: ExternalReference{
				Content: &File{},
			},
		},
		Mode:    "users", // default value
		Uploads: []*Upload{},
	}
}

// Schema returns the JSON schema for the configuration file.
func (conf Config) Schema() ([]byte, error) {
	reflector := jsonschema.Reflector{}
	_ = reflector.AddGoComments("github.com/ctfer-io/ctfd-setup", "./") // this could fail once binary is compiled, thus ignored (no problem)
	r := reflector.Reflect(&Config{})
	r.ID = "https://json.schemastore.org/ctfd.json" // set the Schemastore ID

	return json.MarshalIndent(r, "", "  ")
}

// Validate the configuration content.
func (conf Config) Validate() error {
	// Build schema loader
	schema, err := conf.Schema()
	if err != nil {
		return errors.Wrap(err, "schema validation failed due to schema generation")
	}
	loader := gojsonschema.NewBytesLoader(schema)

	// Load and validate configuration
	confLoader := gojsonschema.NewGoLoader(conf)
	res, err := gojsonschema.Validate(loader, confLoader)
	if err != nil {
		return err
	}
	if !res.Valid() {
		var merr error
		for _, err := range res.Errors() {
			merr = multierr.Append(merr, errors.New(err.String()))
		}
		return merr
	}
	return nil
}
