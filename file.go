package ctfdsetup

import (
	"os"

	"gopkg.in/yaml.v3"
)

type File struct {
	Name    string
	Content []byte
}

var _ yaml.Unmarshaler = (*File)(nil)

func (file *File) UnmarshalYAML(node *yaml.Node) error {
	if node.Value != "" {
		file.Content = []byte(node.Value)
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

	fc, err := os.ReadFile(*lfiv.FromFile)
	if err != nil {
		return err
	}
	file.Content = fc
	return nil
}
