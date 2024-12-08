package ctfdsetup_test

import (
	"testing"

	ctfdsetup "github.com/ctfer-io/ctfd-setup"
	"github.com/stretchr/testify/assert"
)

func Test_U_ConfigSchema(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	cfg := ctfdsetup.NewConfig()

	schema, err := cfg.Schema()
	assert.NoError(err)
	assert.NotEmpty(schema)
}
