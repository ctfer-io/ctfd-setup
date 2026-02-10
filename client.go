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

func apiOptions(ctx context.Context) []api.Option {
	return []api.Option{
		api.WithContext(ctx),
		apiTransport,
	}
}

func GetNonceAndSession(ctx context.Context, url string, opts ...Option) (nonce, session string, err error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return api.GetNonceAndSession(url, apiOptions(ctx)...)
}

type Client struct {
	url string
	sub *api.Client
}

func NewClient(url, nonce, session, apiKey string) *Client {
	return &Client{
		url: url,
		sub: api.NewClient(url, nonce, session, apiKey),
	}
}

func (cli *Client) Bare(ctx context.Context, opts ...Option) (bool, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.url+"/setup", nil)
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

func (cli *Client) Login(ctx context.Context, params *api.LoginParams, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.Login(params, apiOptions(ctx)...)
}

func (cli *Client) Setup(ctx context.Context, params *api.SetupParams, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.Setup(params, apiOptions(ctx)...)
}

// region pages

func (cli *Client) GetPages(ctx context.Context, params *api.GetPagesParams, opts ...Option) ([]*api.Page, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.GetPages(params, apiOptions(ctx)...)
}

func (cli *Client) PatchPage(ctx context.Context, id int, params *api.PatchPageParams, opts ...Option) (*api.Page, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.PatchPage(strconv.Itoa(id), params, apiOptions(ctx)...)
}

func (cli *Client) PostPages(ctx context.Context, params *api.PostPagesParams, opts ...Option) (*api.Page, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.PostPages(params, apiOptions(ctx)...)
}

func (cli *Client) DeletePage(ctx context.Context, id int, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.DeletePage(strconv.Itoa(id), apiOptions(ctx)...)
}

// region files

func (cli *Client) GetFiles(ctx context.Context, params *api.GetFilesParams, opts ...Option) ([]*api.File, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.GetFiles(params, apiOptions(ctx)...)
}

func (cli *Client) PostFiles(ctx context.Context, params *api.PostFilesParams, opts ...Option) ([]*api.File, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.PostFiles(params, apiOptions(ctx)...)
}

// region logos/icons

func (cli *Client) PatchConfigsCTFLogo(ctx context.Context, params *api.PatchConfigsCTFLogo, opts ...Option) (*api.ThemeImage, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.PatchConfigsCTFLogo(params, apiOptions(ctx)...)
}

func (cli *Client) PatchConfigsCTFSmallIcon(ctx context.Context, params *api.PatchConfigsCTFLogo, opts ...Option) (*api.ThemeImage, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.PatchConfigsCTFSmallIcon(params, apiOptions(ctx)...)
}

// region configs

func (cli *Client) PatchConfigs(ctx context.Context, params *api.PatchConfigsParams, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	LogAPICall(ctx)

	return cli.sub.PatchConfigs(params, apiOptions(ctx)...)
}
