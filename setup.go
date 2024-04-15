package ctfdsetup

import (
	"context"
	"net/http"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func Setup(ctx context.Context, url string, conf *Config) error {
	nonce, session, err := api.GetNonceAndSession(url, api.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "getting CTFd nonce and session")
	}
	client := api.NewClient(url, nonce, session, "")

	b, err := bare(ctx, url)
	if err != nil {
		return err
	}
	Log().Info("deciding on CTFd setup strategy", zap.Bool("bare", b))
	if b {
		if err := bareSetup(ctx, client, conf); err != nil {
			return err
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
		return false, err
	}
	return res.StatusCode == 200, nil // 302 if already setup
}

func bareSetup(ctx context.Context, client *api.Client, conf *Config) error {
	Log().Info("setting up fresh CTFd instance")

	logo, err := File(conf.Front.Logo)
	if err != nil {
		return err
	}
	banner, err := File(conf.Front.Banner)
	if err != nil {
		return err
	}
	smallicon, err := File(conf.Front.SmallIcon)
	if err != nil {
		return err
	}

	// Flatten configuration and setup it
	return client.Setup(&api.SetupParams{
		CTFName:                conf.Global.Name,
		CTFDescription:         conf.Global.Description,
		UserMode:               conf.Global.Mode,
		ChallengeVisibility:    conf.Visibilities.Challenge,
		AccountVisibility:      conf.Visibilities.Account,
		ScoreVisibility:        conf.Visibilities.Score,
		RegistrationVisibility: conf.Visibilities.Registration,
		VerifyEmails:           conf.Global.VerifyEmails,
		TeamSize:               conf.Global.TeamSize,
		Name:                   conf.Admin.Name,
		Email:                  conf.Admin.Email,
		Password:               conf.Admin.Password,
		CTFLogo:                logo,
		CTFBanner:              banner,
		CTFSmallIcon:           smallicon,
		CTFTheme:               conf.Front.Theme,
		ThemeColor:             conf.Front.ThemeColor,
		Start:                  conf.Global.Start,
		End:                    conf.Global.End,
	}, api.WithContext(ctx))
}

func updateSetup(_ context.Context, _ *api.Client, _ *Config) error {
	Log().Info("updating existing CTFd instance")

	// TODO implement
	return errors.New("not implemented yet")
}
