package ctfdsetup

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"net/http"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func Setup(ctx context.Context, url string, apiKey string, conf *Config) error {
	ctx, span := Tracer.Start(ctx, "Setup")
	defer span.End()

	nonce, session, err := GetNonceAndSession(ctx, url)
	if err != nil {
		return errors.Wrap(err, "getting CTFd nonce and session")
	}
	client := NewClient(url, nonce, session, apiKey)

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
		if err := client.Login(ctx, &api.LoginParams{
			Name:     conf.Admin.Name,
			Password: conf.Admin.Password.Content,
		}); err != nil {
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
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	res, err := client.Do(req)
	if err != nil {
		return false, &ErrClient{err: err}
	}
	return res.StatusCode == 200, nil // 302 if already setup
}

func bareSetup(ctx context.Context, client *Client, conf *Config) error {
	// Flatten configuration and (basic) setup it
	if err := client.Setup(ctx, &api.SetupParams{
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
	}); err != nil {
		return &ErrClient{err: err}
	}
	return nil
}

func updateSetup(ctx context.Context, client *Client, conf *Config) error {
	// Push logo
	if conf.Theme.Logo.Name != "" {
		lf, err := client.PostFiles(ctx, &api.PostFilesParams{
			Files: []*api.InputFile{
				(*api.InputFile)(conf.Theme.Logo),
			},
		})
		if err != nil {
			return errors.Wrap(err, "pushing theme logo")
		}
		if _, err := client.PatchConfigsCTFLogo(ctx, &api.PatchConfigsCTFLogo{
			Value: &lf[0].Location,
		}); err != nil {
			return errors.Wrap(err, "patching CTF logo")
		}
	} else {
		if _, err := client.PatchConfigsCTFLogo(ctx, &api.PatchConfigsCTFLogo{}); err != nil {
			return err
		}
	}
	// TODO else delete logo

	// Push small icon
	if conf.Theme.SmallIcon.Name != "" {
		smf, err := client.PostFiles(ctx, &api.PostFilesParams{
			Files: []*api.InputFile{
				(*api.InputFile)(conf.Theme.SmallIcon),
			},
		})
		if err != nil {
			return errors.Wrap(err, "pushing theme small icon")
		}
		if _, err := client.PatchConfigsCTFSmallIcon(ctx, &api.PatchConfigsCTFLogo{
			Value: &smf[0].Location,
		}); err != nil {
			return errors.Wrap(err, "patching CTF small icon")
		}
	} else {
		if _, err := client.PatchConfigsCTFSmallIcon(ctx, &api.PatchConfigsCTFLogo{}); err != nil {
			return err
		}
	}
	// TODO else delete small icon

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
		DomainBlacklist:                    conf.Accounts.DomainBlacklist,
		IncorrectSubmissionsPerMin:         conf.Accounts.IncorrectSubmissionsPerMinute,
		NameChanges:                        conf.Accounts.NameChanges,
		NumTeams:                           conf.Accounts.NumTeams,
		NumUsers:                           conf.Accounts.NumUsers,
		TeamCreation:                       conf.Accounts.TeamCreation,
		TeamDisbanding:                     conf.Accounts.TeamDisbanding,
		TeamSize:                           conf.Accounts.TeamSize,
		PasswordMinLength:                  conf.Accounts.PasswordMinLength,
		VerifyEmails:                       &conf.Accounts.VerifyEmails,
		ViewSelfSubmission:                 conf.Challenges.ViewSelfSubmission,
		MaxAttemptsBehavior:                conf.Challenges.MaxAttemptsBehavior,
		MaxAttemptsTimeout:                 conf.Challenges.MaxAttemptsTimeout,
		HintsFreePublicAccess:              conf.Challenges.HintsFreePublicAccess,
		ChallengeRatings:                   conf.Challenges.ChallengeRatings,
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
		SocialSharesTemplate:               string(conf.Social.Template.Content),
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

	if err := client.PatchConfigs(ctx, params); err != nil {
		return &ErrClient{err: err}
	}

	// Handle additional pages configuration
	if conf.Pages != nil && len(conf.Pages.Additional) != 0 {
		if err := additionalPages(ctx, client, conf.Pages.Additional); err != nil {
			return err
		}
	}

	// Upload files
	if len(conf.Uploads) != 0 {
		var merr error
		for _, f := range conf.Uploads {
			// Compute file hash
			h := sha1.New()
			_, err := h.Write(f.File.Content)
			if err != nil {
				merr = multierr.Append(merr, errors.Wrapf(err, "computing hash of %s", f.File.Name))
				continue
			}
			x := hex.EncodeToString(h.Sum(nil))

			// Get the file from CTFd
			fs, err := client.GetFiles(ctx, &api.GetFilesParams{
				Location: &f.Location,
			})
			if err != nil {
				merr = multierr.Append(merr, errors.Wrapf(err, "getting file at %s", f.Location))
				continue
			}

			// Check if need re-push
			if len(fs) != 0 && fs[0].SHA1sum == x {
				continue
			}

			// Else push it (or update it)
			logger.Debug("uploading file",
				zap.String("location", f.Location),
			)
			if _, err := client.PostFiles(ctx, &api.PostFilesParams{
				Files: []*api.InputFile{
					(*api.InputFile)(f.File),
				},
				Location: &f.Location,
			}); err != nil {
				merr = multierr.Append(merr, err)
			}
		}
		if merr != nil {
			return merr
		}
	}

	return nil
}

func ptr[T any](t T) *T {
	return &t
}
