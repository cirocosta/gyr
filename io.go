package gyr

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func WriteYAML(w io.Writer, nodes []*yaml.Node) error {
	const YAMLDocumentSeparator = "---\n"

	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	for _, node := range nodes {
		if err := encoder.Encode(node); err != nil {
			return fmt.Errorf("encode: %w", err)
		}
	}

	return nil
}

func NodesFromFiles(fnames []string) ([]*yaml.Node, error) {
	res := []*yaml.Node{}

	for _, fname := range fnames {
		nodes, err := NodesFromFile(fname)
		if err != nil {
			return nil, err
		}

		res = append(res, nodes...)
	}

	return res, nil
}

func NodesFromFile(fname string) ([]*yaml.Node, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("open '%s': %w", fname, err)
	}
	defer f.Close()

	return NodesFromReader(f)
}

func NodesFromReader(reader io.Reader) ([]*yaml.Node, error) {
	objs := []*yaml.Node{}
	decoder := yaml.NewDecoder(reader)

	for {
		obj := &yaml.Node{}

		err := decoder.Decode(obj)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf(
				"decode into monero nodeset: %w",
				err,
			)
		}

		objs = append(objs, obj)
	}

	return objs, nil
}
