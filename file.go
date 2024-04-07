package ctfdsetup

import (
	"os"

	"github.com/ctfer-io/go-ctfd/api"
	"gopkg.in/yaml.v3"
)

type File struct {
	Name    string
	Content []byte
}

// Make it able to be unmarshalled from YAML
var _ yaml.Unmarshaler = (*Secret)(nil)

func (file *File) UnmarshalYAML(node *yaml.Node) error {
	f, err := os.ReadFile(node.Value)
	if err != nil {
		return err
	}
	*file = File{
		Name:    node.Value,
		Content: f,
	}
	return nil
}

func (file File) ToInputFile() *api.InputFile {
	if file.Name == "" { // file is not set
		return nil
	}

	f := api.InputFile(file)
	return &f
}
