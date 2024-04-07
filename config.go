package ctfdsetup

import "fmt"

const (
	DefaultConf = `
ctf_name:                "" # must be set
ctf_description:         "" # must be set
user_mode:               users
visibilities:
  challenge:    public
  account:      public
  score:        public
  registration: public
verify_emails:           true
theme: core
start: ""
end:   ""
# don't set team_size, logo, banner and small_icon by default

admin:
  name:
    from_env: CTFD_ADMIN_NAME      # default admin name is fetched from this environment variable
  email:
    from_env: CTFD_ADMIN_EMAIL     # default admin email is fetched from this environment variable
  password:
    from_env: CTFD_ADMIN_PASSWORD  # default admin password is fetched from this environment variable
`
)

type Conf struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	UserMode     string `yaml:"user_mode"`
	Visibilities struct {
		Challenge    string `yaml:"challenge"`
		Account      string `yaml:"account"`
		Score        string `yaml:"score"`
		Registration string `yaml:"registration"`
	} `yaml:"visibilities"`
	VerifyEmails bool   `yaml:"verify_emails"`
	CTFTheme     string `yaml:"theme"`
	ThemeColor   string `yaml:"theme_color"`
	Start        string `yaml:"start"`
	End          string `yaml:"end"`
	TeamSize     *int   `yaml:"team_size,omitempty"`
	CTFLogo      File   `yaml:"logo"`
	CTFBanner    File   `yaml:"banner"`
	CTFSmallIcon File   `yaml:"small_icon"`
	Admin        struct {
		Name     Secret `yaml:"name"`
		Email    Secret `yaml:"email"`
		Password Secret `yaml:"password"`
	} `yaml:"admin"`
}

func (conf Conf) Validate() error {
	if conf.Name == "" {
		return ErrInvalidConf("name")
	}
	if conf.Description == "" {
		return ErrInvalidConf("description")
	}
	if conf.Admin.Name == "" {
		return ErrInvalidConf("admin.name")
	}
	if conf.Admin.Email == "" {
		return ErrInvalidConf("admin.email")
	}
	if conf.Admin.Password == "" {
		return ErrInvalidConf("admin.password")
	}
	return nil
}

type ErrInvalidConf string

var _ error = (*ErrInvalidConf)(nil)

func (err ErrInvalidConf) Error() string {
	return fmt.Sprintf("%s is invalid", string(err))
}
