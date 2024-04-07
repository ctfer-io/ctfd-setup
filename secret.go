package ctfdsetup

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Secret string

// Ensure secret don't get leaked in case of standard print
var _ fmt.Stringer = (*Secret)(nil)

func (sec Secret) String() string {
	return "[REDACTED]"
}

func (sec Secret) ToString() string {
	return string(sec)
}

// Make it able to be unmarshalled from YAML
var _ yaml.Unmarshaler = (*Secret)(nil)

func (sec *Secret) UnmarshalYAML(node *yaml.Node) error {
	// If hardcoded (bad practice) keep it
	if node.Value != "" {
		*sec = Secret(node.Value)
		return nil
	}

	// Get it from either env or file
	type sources struct {
		FromEnv  *string `yaml:"from_env,omitempty"`
		FromFile *string `yaml:"from_file,omitempty"`
	}
	var srcs sources
	if err := node.Decode(&srcs); err != nil {
		return err
	}
	switch {
	case srcs.FromEnv != nil:
		*sec = Secret(os.Getenv(*srcs.FromEnv))
	case srcs.FromFile != nil:
		b, err := os.ReadFile(*srcs.FromFile)
		if err != nil {
			return err
		}
		*sec = Secret(string(b))
	default:
		return errors.New("invalid syntax")
	}
	return nil
}
