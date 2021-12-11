package internal

import (
	"os"

	"github.com/airfocusio/go-expandenv"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Config struct {
	Values map[string]string `yaml:"values"`
}

func Run() error {
	config := new(Config)
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			yamlIn := items[i].YNode()
			bytesIn, err := yaml.Marshal(yamlIn)
			if err != nil {
				return nil, err
			}

			var yamlRaw interface{}
			err = yaml.Unmarshal(bytesIn, &yamlRaw)
			if err != nil {
				return nil, err
			}
			yamlRaw, err = expandenv.Expand(yamlRaw, config.Values)
			if err != nil {
				return nil, err
			}
			bytesOut, err := yaml.Marshal(yamlRaw)
			if err != nil {
				return nil, err
			}

			yamlOut := yaml.Node{}
			err = yaml.Unmarshal(bytesOut, &yamlOut)
			if err != nil {
				return nil, err
			}
			items[i].SetYNode(&yamlOut)
		}

		return items, nil
	}

	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	err := cmd.Execute()
	if err != nil {
		os.Stderr.WriteString("\n")
		return err
	}
	return nil
}
