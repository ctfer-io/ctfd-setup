package ctfdsetup

import "github.com/pkg/errors"

type ErrClient struct {
	err error
}

var _ error = (*ErrClient)(nil)

func (err ErrClient) Error() string {
	return errors.Wrap(err.err, "client error").Error()
}
