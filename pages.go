package ctfdsetup

import (
	"context"
	"slices"

	"github.com/ctfer-io/go-ctfd/api"
)

func additionalPages(ctx context.Context, client *Client, pages []Page) error {
	ctfdPages, err := client.GetPages(ctx, nil)
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
			if _, err := client.PatchPage(ctx, ctfdP.ID, &api.PatchPageParams{
				Title:        page.Title,
				Route:        page.Route,
				Format:       page.Format,
				Content:      string(page.Content.Content),
				Draft:        page.Draft,
				Hidden:       page.Hidden,
				AuthRequired: page.AuthRequired,
			}); err != nil {
				return err
			}
		} else {
			// CREATE
			if _, err := client.PostPages(ctx, &api.PostPagesParams{
				Title:        page.Title,
				Route:        page.Route,
				Format:       page.Format,
				Content:      string(page.Content.Content),
				Draft:        page.Draft,
				Hidden:       page.Hidden,
				AuthRequired: page.AuthRequired,
			}); err != nil {
				return err
			}
		}
	}

	// DELETE
	for _, ctfdP := range ctfdPages {
		if !slices.Contains(cu, ctfdP.Route) {
			if err := client.DeletePage(ctx, ctfdP.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
