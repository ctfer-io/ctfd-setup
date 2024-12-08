package ctfdsetup

import (
	"context"
	"slices"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
)

func additionalPages(ctx context.Context, client *api.Client, pages []Page) error {
	ctfdPages, err := client.GetPages(&api.GetPagesParams{}, api.WithContext(ctx))
	if err != nil {
		return err
	}
	cu := []string{}

	for _, page := range pages {
		var ctfdP *api.Page
		for _, p := range ctfdPages {
			if p.Route == page.Route {
				ctfdP = p
				break
			}
		}

		cu = append(cu, page.Route)
		if ctfdP != nil {
			// UPDATE
			if _, err := client.PatchPage(strconv.Itoa(ctfdP.ID), &api.PatchPageParams{
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
			// CREATE
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

	// DELETE
	for _, ctfdP := range ctfdPages {
		if !slices.Contains(cu, ctfdP.Route) {
			if err := client.DeletePage(strconv.Itoa(ctfdP.ID)); err != nil {
				return err
			}
		}
	}
	return nil
}
