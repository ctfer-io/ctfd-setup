// Contains a wrapper around github.com/ctfer-io/go-ctfd.
//
// It injects spans for all API operations.

package ctfdsetup

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var apiTransport = api.WithTransport(otelhttp.NewTransport(http.DefaultTransport))

func options(ctx context.Context) []api.Option {
	return []api.Option{
		api.WithContext(ctx),
		apiTransport,
	}
}

func GetNonceAndSession(ctx context.Context, url string) (nonce, session string, err error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return api.GetNonceAndSession(url, options(ctx)...)
}

type Client struct {
	sub *api.Client
}

func NewClient(url, nonce, session, apiKey string) *Client {
	return &Client{
		sub: api.NewClient(url, nonce, session, apiKey),
	}
}

func (cli *Client) Login(ctx context.Context, params *api.LoginParams) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.Login(params, options(ctx)...)
}

func (cli *Client) Setup(ctx context.Context, params *api.SetupParams) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.Setup(params, options(ctx)...)
}

// region pages

func (cli *Client) GetPages(ctx context.Context, params *api.GetPagesParams) ([]*api.Page, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetPages(params, options(ctx)...)
}

func (cli *Client) PatchPage(ctx context.Context, id int, params *api.PatchPageParams) (*api.Page, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchPage(strconv.Itoa(id), params, options(ctx)...)
}

func (cli *Client) PostPages(ctx context.Context, params *api.PostPagesParams) (*api.Page, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostPages(params, options(ctx)...)
}

func (cli *Client) DeletePage(ctx context.Context, id int) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeletePage(strconv.Itoa(id), options(ctx)...)
}

// region files

func (cli *Client) GetFiles(ctx context.Context, params *api.GetFilesParams) ([]*api.File, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetFiles(params, options(ctx)...)
}

func (cli *Client) PostFiles(ctx context.Context, params *api.PostFilesParams) ([]*api.File, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostFiles(params, options(ctx)...)
}

// region logos/icons

func (cli *Client) PatchConfigsCTFLogo(ctx context.Context, params *api.PatchConfigsCTFLogo) (*api.ThemeImage, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchConfigsCTFLogo(params, options(ctx)...)
}

func (cli *Client) PatchConfigsCTFSmallIcon(ctx context.Context, params *api.PatchConfigsCTFLogo) (*api.ThemeImage, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchConfigsCTFSmallIcon(params, options(ctx)...)
}

// region configs

func (cli *Client) PatchConfigs(ctx context.Context, params *api.PatchConfigsParams) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchConfigs(params, options(ctx)...)
}
