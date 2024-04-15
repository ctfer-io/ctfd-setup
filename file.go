package ctfdsetup

import (
	"os"
	"path"

	"github.com/ctfer-io/go-ctfd/api"
)

func File(loc *string) (*api.InputFile, error) {
	if loc == nil || *loc == "" {
		return nil, nil
	}
	b, err := os.ReadFile(*loc)
	if err != nil {
		return nil, err
	}
	return &api.InputFile{
		Name:    path.Base(*loc),
		Content: b,
	}, nil
}
