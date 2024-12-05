package ctfdsetup_test

import (
	"testing"

	ctfdsetup "github.com/ctfer-io/ctfd-setup"
	"github.com/stretchr/testify/assert"
)

func Test_U_ConfigSchema(t *testing.T) {
	assert := assert.New(t)

	cfg := &ctfdsetup.Config{}
	schema, err := cfg.Schema()
	assert.NoError(err)
	assert.NotEmpty(schema)
}
