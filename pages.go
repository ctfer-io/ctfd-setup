package ctfdsetup

import (
	"context"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
)

func additionalPages(ctx context.Context, client *api.Client, pages []Page) error {
	for _, page := range pages {
		ctfdP, err := client.GetPages(&api.GetPagesParams{
			Route: ptr(page.Route),
		}, api.WithContext(ctx))
		if err != nil {
			return err
		}

		exist := len(ctfdP) == 1
		if exist {
			if _, err := client.PatchPage(strconv.Itoa(ctfdP[0].ID), &api.PatchPageParams{
				Title:        page.Title,
				Route:        page.Route,
				Format:       page.Format,
				Content:      string(page.Content.Content),
				Draft:        page.Draft,
				Hidden:       page.Hidden,
				AuthRequired: page.AuthRequired,
			}, api.WithContext(ctx)); err != nil {
				return err
			}
		} else {
			if _, err := client.PostPages(&api.PostPagesParams{
				Title:        page.Title,
				Route:        page.Route,
				Format:       page.Format,
				Content:      string(page.Content.Content),
				Draft:        page.Draft,
				Hidden:       page.Hidden,
				AuthRequired: page.AuthRequired,
			}, api.WithContext(ctx)); err != nil {
				return err
			}
		}
	}
	return nil
}
