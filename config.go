package ctfdsetup

import (
	"fmt"

	"go.uber.org/multierr"
)

type (
	Config struct {
		Global       Global       `yaml:"global"`
		Visibilities Visibilities `yaml:"visibilities"`
		Front        Front        `yaml:"front"`
		Admin        Admin        `yaml:"admin"`
	}

	Global struct {
		Name         string `yaml:"name"`        // required
		Description  string `yaml:"description"` // required
		Mode         string `yaml:"mode"`
		TeamSize     *int   `yaml:"team_size"`
		VerifyEmails bool   `yaml:"verify_emails"`
		Start        string `yaml:"start"`
		End          string `yaml:"end"`
	}

	Visibilities struct {
		Challenge    string `yaml:"challenge"`
		Account      string `yaml:"account"`
		Score        string `yaml:"score"`
		Registration string `yaml:"registration"`
	}

	Front struct {
		Theme      string  `yaml:"theme"`
		ThemeColor string  `yaml:"theme_color"`
		Logo       *string `yaml:"logo"`
		Banner     *string `yaml:"banner"`
		SmallIcon  *string `yaml:"small_icon"`
	}

	Admin struct {
		Name     string `yaml:"name"`     // required
		Email    string `yaml:"email"`    // required
		Password string `yaml:"password"` // required
	}
)

func (conf Config) Validate() error {
	var merr error
	if conf.Global.Name == "" {
		merr = multierr.Append(merr, &ErrRequired{Attribute: "global.name"})
	}
	if conf.Global.Description == "" {
		merr = multierr.Append(merr, &ErrRequired{Attribute: "global.description"})
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
	// that and provide a meaningful error message

	return nil
}

type ErrRequired struct {
	Attribute string
}

var _ error = (*ErrRequired)(nil)

func (err ErrRequired) Error() string {
	return fmt.Sprintf("Required attribute %s was either not set or left to empty value", err.Attribute)
}
