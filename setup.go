package ctfdsetup

import (
	"context"
	"net/http"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func Setup(ctx context.Context, url string, apiKey string, conf *Config) error {
	nonce, session, err := api.GetNonceAndSession(url, api.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "getting CTFd nonce and session")
	}
	client := api.NewClient(url, nonce, session, apiKey)

	b, err := bare(ctx, url)
	if err != nil {
		return err
	}
	Log().Info("deciding on CTFd setup strategy",
		zap.Bool("bare", b),
		zap.Bool("login", apiKey == ""),
	)
	if b {
		if err := bareSetup(ctx, client, conf); err != nil {
			return err
		}
	} else if apiKey == "" {
		if err := client.Login(&api.LoginParams{
			Name:     conf.Admin.Name,
			Password: conf.Admin.Password.Content,
		}, api.WithContext(ctx)); err != nil {
			return &ErrClient{err: err}
		}
	}
	return updateSetup(ctx, client, conf)
}

func bare(ctx context.Context, url string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/setup", nil)
	if err != nil {
		return false, err
	}
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return false, &ErrClient{err: err}
	}
	return res.StatusCode == 200, nil // 302 if already setup
}

func bareSetup(ctx context.Context, client *api.Client, conf *Config) error {
	// Flatten configuration and (basic) setup it
	if err := client.Setup(&api.SetupParams{
		CTFName:                conf.Appearance.Name,
		CTFDescription:         conf.Appearance.Description,
		UserMode:               conf.Mode,
		CTFTheme:               conf.Theme.Name,
		ChallengeVisibility:    conf.Settings.ChallengeVisibility,
		AccountVisibility:      conf.Settings.AccountVisibility,
		ScoreVisibility:        conf.Settings.ScoreVisibility,
		RegistrationVisibility: conf.Settings.RegistrationVisibility,
		VerifyEmails:           conf.Accounts.VerifyEmails,
		TeamSize:               conf.Accounts.TeamSize,
		Name:                   conf.Admin.Name,
		Email:                  conf.Admin.Email,
		Password:               conf.Admin.Password.Content,
	}, api.WithContext(ctx)); err != nil {
		return &ErrClient{err: err}
	}
	return nil
}

func updateSetup(ctx context.Context, client *api.Client, conf *Config) error {
	// Push logo
	if conf.Theme.Logo.Name != "" {
		lf, err := client.PostFiles(&api.PostFilesParams{
			Files: []*api.InputFile{
				(*api.InputFile)(conf.Theme.Logo),
			},
		}, api.WithContext(ctx))
		if err != nil {
			return errors.Wrap(err, "pushing theme logo")
		}
		if _, err := client.PatchConfigsCTFLogo(&api.PatchConfigsCTFLogo{
			Value: &lf[0].Location,
		}, api.WithContext(ctx)); err != nil {
			return errors.Wrap(err, "patching CTF logo")
		}
	}

	// Push small icon
	if conf.Theme.SmallIcon.Name != "" {
		smf, err := client.PostFiles(&api.PostFilesParams{
			Files: []*api.InputFile{
				(*api.InputFile)(conf.Theme.SmallIcon),
			},
		}, api.WithContext(ctx))
		if err != nil {
			return errors.Wrap(err, "pushing theme small icon")
		}
		if _, err := client.PatchConfigsCTFSmallIcon(&api.PatchConfigsCTFLogo{
			Value: &smf[0].Location,
		}, api.WithContext(ctx)); err != nil {
			return errors.Wrap(err, "patching CTF small icon")
		}
	}

	// Update configs attributes
	params := &api.PatchConfigsParams{
		CTFDescription:                     &conf.Appearance.Description,
		CTFName:                            &conf.Appearance.Name,
		DefaultLocale:                      conf.Appearance.DefaultLocale,
		CTFTheme:                           &conf.Theme.Name,
		ThemeFooter:                        ptr(string(conf.Theme.Footer.Content)),
		ThemeHeader:                        ptr(string(conf.Theme.Header.Content)),
		ThemeSettings:                      ptr(string(conf.Theme.Settings.Content)),
		DomainWhitelist:                    conf.Accounts.DomainWhitelist,
		IncorrectSubmissionsPerMin:         conf.Accounts.IncorrectSubmissionsPerMinute,
		NameChanges:                        conf.Accounts.NameChanges,
		NumTeams:                           conf.Accounts.NumTeams,
		NumUsers:                           conf.Accounts.NumUsers,
		TeamCreation:                       conf.Accounts.TeamCreation,
		TeamDisbanding:                     conf.Accounts.TeamDisbanding,
		TeamSize:                           conf.Accounts.TeamSize,
		VerifyEmails:                       &conf.Accounts.VerifyEmails,
		RobotsTxt:                          ptr(string(conf.Pages.RobotsTxt.Content)),
		OauthClientID:                      conf.MajorLeagueCyber.ClientID,
		OauthClientSecret:                  conf.MajorLeagueCyber.ClientSecret,
		AccountVisibility:                  &conf.Settings.AccountVisibility,
		ChallengeVisibility:                &conf.Settings.ChallengeVisibility,
		RegistrationVisibility:             &conf.Settings.RegistrationVisibility,
		ScoreVisibility:                    &conf.Settings.ScoreVisibility,
		Paused:                             conf.Settings.Paused,
		HTMLSanitization:                   conf.Security.HTMLSanitization,
		RegistrationCode:                   conf.Security.RegistrationCode,
		MailUseAuth:                        nil, // Handled later
		MailUsername:                       nil, // Handled later
		MailPassword:                       nil, // Handled later
		MailFromAddr:                       nil, // Deprecated, set to nil for autocomplete
		MailGunAPIKey:                      nil, // Deprecated, set to nil for autocomplete
		MailGunBaseURL:                     nil, // Deprecated, set to nil for autocomplete
		MailPort:                           conf.Email.Port,
		MailServer:                         conf.Email.Server,
		MailSSL:                            conf.Email.TLS_SSL,
		MailTLS:                            conf.Email.STARTTLS,
		SuccessfulRegistrationEmailSubject: conf.Email.Registration.Subject,
		SuccessfulRegistrationEmailBody:    conf.Email.Registration.Body,
		VerificationEmailSubject:           conf.Email.Confirmation.Subject,
		VerificationEmailBody:              conf.Email.Confirmation.Body,
		UserCreationEmailSubject:           conf.Email.NewAccount.Subject,
		UserCreationEmailBody:              conf.Email.NewAccount.Body,
		PasswordChangeAlertSubject:         conf.Email.PasswordReset.Subject,
		PasswordChangeAlertBody:            conf.Email.PasswordReset.Body,
		PasswordResetSubject:               conf.Email.PasswordResetConfirmation.Subject,
		PasswordResetBody:                  conf.Email.PasswordResetConfirmation.Body,
		Start:                              conf.Time.Start,
		End:                                conf.Time.End,
		Freeze:                             conf.Time.Freeze,
		ViewAfterCTF:                       conf.Time.ViewAfter,
		SocialShares:                       conf.Social.Shares,
		PrivacyURL:                         conf.Legal.PrivacyPolicy.URL,
		PrivacyText:                        ptr(string(conf.Legal.PrivacyPolicy.Content.Content)),
		TOSURL:                             conf.Legal.TOS.URL,
		TOSText:                            ptr(string(conf.Legal.TOS.Content.Content)),
		UserMode:                           &conf.Mode,
	}

	// Handle mail server authentication
	if conf.Email.Username != nil && conf.Email.Password != nil {
		params.MailUseAuth = ptr(true)
		params.MailUsername = conf.Email.Username
		params.MailPassword = conf.Email.Password
	}

	if err := client.PatchConfigs(params, api.WithContext(ctx)); err != nil {
		return &ErrClient{err: err}
	}
	return nil
}

func ptr[T any](t T) *T {
	return &t
}
