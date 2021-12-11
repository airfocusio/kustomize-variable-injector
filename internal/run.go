package internal

import (
	"os"
	"strings"

	"github.com/airfocusio/go-expandenv"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Target struct {
	Group     string `yaml:"group"`
	Version   string `yaml:"version"`
	Kind      string `yaml:"kind"`
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type Config struct {
	Replacements []struct {
		Target    Target            `yaml:"target"`
		Variables map[string]string `yaml:"variables"`
	} `yaml:"replacements"`
}

func Run() error {
	config := new(Config)
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			match := false
			variables := map[string]string{}
			for _, t := range config.Replacements {
				group, version := parseApiVersion(items[i].GetApiVersion())
				kind := items[i].GetKind()
				name := items[i].GetName()
				namespace := items[i].GetNamespace()
				if (t.Target.Group == "" || t.Target.Group == group) &&
					(t.Target.Version == "" || t.Target.Version == version) &&
					(t.Target.Kind == "" || t.Target.Kind == kind) &&
					(t.Target.Name == "" || t.Target.Name == name) &&
					(t.Target.Namespace == "" || t.Target.Namespace == namespace) {
					for k, v := range t.Variables {
						variables[k] = v
					}
					match = true
				}
			}

			if match {
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
				yamlRaw, err = expandenv.Expand(yamlRaw, variables)
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

func parseApiVersion(apiVersion string) (string, string) {
	splitted := strings.SplitN(apiVersion, "/", 2)
	if len(splitted) == 2 {
		return splitted[0], splitted[1]
	}
	return "", splitted[0]
}
