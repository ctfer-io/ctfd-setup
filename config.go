package ctfdsetup

import (
	"fmt"

	"go.uber.org/multierr"
)

type (
	Config struct {
		Appearance Appearance `yaml:"appearance"`
		Theme      Theme      `yaml:"theme"`
		Accounts   Accounts   `yaml:"accounts"`
		Pages      Pages      `yaml:"pages"`
		// Don't handle brackets here, should not be part of those settings but CRUD objects
		// CustomFields are not handled as they are not predictable and would be hard to handle + bad practice (API changes on the fly)
		MajorLeagueCyber MajorLeagueCyber `yaml:"major_league_cyber"`
		Settings         Settings         `yaml:"settings"`
		Security         Security         `yaml:"security"`
		Email            Email            `yaml:"email"`
		Time             Time             `yaml:"time"`
		Social           Social           `yaml:"social"`
		Legal            Legal            `yaml:"legal"`
		Mode             string           `yaml:"mode"`

		Admin Admin `yaml:"admin"`
	}

	Appearance struct {
		Name          string  `yaml:"name"`        // required
		Description   string  `yaml:"description"` // required
		DefaultLocale *string `yaml:"default_locale"`
	}

	Theme struct {
		Logo      *File  `yaml:"logo"`
		SmallIcon *File  `yaml:"small_icon"`
		Name      string `yaml:"name"`
		Color     string `yaml:"color"`
		// Banner is only supported by bare setup, need to be at least support by PatchConfigs
		Header   *File `yaml:"header"`
		Footer   *File `yaml:"footer"`
		Settings *File `yaml:"settings"`
	}

	Accounts struct {
		DomainWhitelist               *string `yaml:"domain_whitelist"`
		VerifyEmails                  bool    `yaml:"verify_emails"`
		TeamCreation                  *bool   `yaml:"team_creation"`
		TeamSize                      *int    `yaml:"team_size"`
		NumTeams                      *int    `yaml:"num_teams"`
		NumUsers                      *int    `yaml:"num_users"`
		TeamDisbanding                *string `yaml:"team_disbanding"`
		IncorrectSubmissionsPerMinute *int    `yaml:"incorrect_submissions_per_minutes"`
		NameChanges                   *bool   `yaml:"name_changes"`
	}

	Pages struct {
		RobotsTxt *File `yaml:"robots_txt"`
	}

	MajorLeagueCyber struct {
		ClientID     *string `yaml:"client_id"`
		ClientSecret *string `yaml:"client_secret"`
	}

	Settings struct {
		ChallengeVisibility    string `yaml:"challenge_visibility"`
		AccountVisibility      string `yaml:"account_visibility"`
		ScoreVisibility        string `yaml:"score_visibility"`
		RegistrationVisibility string `yaml:"registration_visibility"`
		Paused                 *bool  `yaml:"paused"`
	}

	Security struct {
		HTMLSanitization *bool   `yaml:"html_sanitization"`
		RegistrationCode *string `yaml:"registration_code"`
	}

	Email struct {
		Registration              EmailContent `yaml:"registration"`
		Confirmation              EmailContent `yaml:"confirmation"`
		NewAccount                EmailContent `yaml:"new_account"`
		PasswordReset             EmailContent `yaml:"password_reset"`
		PasswordResetConfirmation EmailContent `yaml:"password_reset_confirmation"`
		From                      *string      `yaml:"from"`
		Server                    *string      `yaml:"server"`
		Port                      *string      `yaml:"port"`
		Username                  *string      `yaml:"username"`
		Password                  *string      `yaml:"password"`
		TLS_SSL                   *bool        `yaml:"tls_ssl"`
		STARTTLS                  *bool        `yaml:"starttls"`
	}

	EmailContent struct {
		Subject *string `yaml:"subject"`
		Body    *string `yaml:"body"`
	}

	Time struct {
		Start     *string `yaml:"start"`
		End       *string `yaml:"end"`
		Freeze    *string `yaml:"freeze"`
		ViewAfter *bool   `yaml:"view_after"`
	}

	Social struct {
		Shares *bool `yaml:"shares"`
	}

	Legal struct {
		TOS           ExternalReference `yaml:"tos"`
		PrivacyPolicy ExternalReference `yaml:"privacy_policy"`
	}

	ExternalReference struct {
		URL     *string `yaml:"url"`
		Content *string `yaml:"content"`
	}

	Admin struct {
		Name     string `yaml:"name"`     // required
		Email    string `yaml:"email"`    // required
		Password string `yaml:"password"` // required
	}
)

func (conf Config) Validate() error {
	var merr error
	if conf.Appearance.Name == "" {
		merr = multierr.Append(merr, &ErrRequired{Attribute: "appearance.name"})
	}
	if conf.Appearance.Description == "" {
		merr = multierr.Append(merr, &ErrRequired{Attribute: "appearance.description"})
	}
	if conf.Admin.Name == "" {
		merr = multierr.Append(merr, &ErrRequired{Attribute: "admin.name"})
	}
	if conf.Admin.Email == "" {
		merr = multierr.Append(merr, &ErrRequired{Attribute: "admin.email"})
	}
	if conf.Admin.Password == "" {
		merr = multierr.Append(merr, &ErrRequired{Attribute: "admin.password"})
	}
	if merr != nil {
		return merr
	}

	// Does not validate attributes content, let CTFd deal with
	// that and provide a meaningful error message... if it can :)

	return nil
}

type ErrRequired struct {
	Attribute string
}

var _ error = (*ErrRequired)(nil)

func (err ErrRequired) Error() string {
	return fmt.Sprintf("Required attribute %s was either not set or left to empty value", err.Attribute)
}
