package ctfdsetup

import (
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"gopkg.in/yaml.v3"
)

var (
	Directory string
)

type File struct {
	Name    string `yaml:"-" json:"-" jsonschema:"-"`
	Content []byte `yaml:"-" json:"-" jsonschema:"-"`
}

var _ yaml.Unmarshaler = (*File)(nil)

func (file *File) UnmarshalYAML(node *yaml.Node) error {
	if node.Value != "" {
		file.Content = []byte(node.Value)
		return nil
	}
	type lfi struct {
		FromFile *string `yaml:"from_file"`
	}
	var lfiv lfi
	if err := node.Decode(&lfiv); err != nil {
		return err
	}

	if lfiv.FromFile == nil {
		return nil
	}

	fc, err := os.ReadFile(filepath.Join(Directory, *lfiv.FromFile))
	if err != nil {
		return err
	}
	file.Name = *lfiv.FromFile
	file.Content = fc
	return nil
}

func (file File) JSONSchema() *jsonschema.Schema {
	subObj := jsonschema.NewProperties()
	subObj.Set("from_file", &jsonschema.Schema{
		Type:        "string",
		Description: "The file to import content from.",
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
