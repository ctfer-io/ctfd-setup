package ctfdsetup

import (
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"gopkg.in/yaml.v3"
)

type FromEnv struct {
	Content string `yaml:"-" json:"-" jsonschema:"-"`
}

var _ yaml.Unmarshaler = (*FromEnv)(nil)

func (fe *FromEnv) UnmarshalYAML(node *yaml.Node) error {
	if node.Value != "" {
		fe.Content = node.Value
		return nil
	}
	type lfe struct {
		FromEnv *string `yaml:"from_env"`
	}
	var lfev lfe
	if err := node.Decode(&lfev); err != nil {
		return err
	}

	if lfev.FromEnv == nil {
		return nil
	}

	fe.Content = os.Getenv(*lfev.FromEnv)
	if len(fe.Content) == 0 {
		return fmt.Errorf("empty value from environment variable %s", *lfev.FromEnv)
	}
	return nil
}

func (fe FromEnv) JSONSchema() *jsonschema.Schema {
	subObj := jsonschema.NewProperties()
	subObj.Set("from_env", &jsonschema.Schema{
		Type:        "string",
		Description: "The environment variable to look at",
	})

	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type:       "object",
				Properties: subObj,
			}, {
				Type: "string",
			},
		},
	}
}
