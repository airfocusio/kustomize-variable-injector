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

type Replacement struct {
	Targets   []Target          `yaml:"targets"`
	Variables map[string]string `yaml:"variables"`
}

type Config struct {
	Replacements []Replacement `yaml:"replacements"`
}

func Run() error {
	config := new(Config)
	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(replace(config))}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	err := cmd.Execute()
	if err != nil {
		os.Stderr.WriteString("\n")
		return err
	}
	return nil
}

func replace(config *Config) func(items []*yaml.RNode) ([]*yaml.RNode, error) {
	return func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			match := false
			variables := map[string]string{}
			for _, r := range config.Replacements {
				group, version := parseApiVersion(items[i].GetApiVersion())
				kind := items[i].GetKind()
				name := items[i].GetName()
				namespace := items[i].GetNamespace()
				for _, t := range r.Targets {
					if (t.Group == "" || t.Group == group) &&
						(t.Version == "" || t.Version == version) &&
						(t.Kind == "" || t.Kind == kind) &&
						(t.Name == "" || t.Name == name) &&
						(t.Namespace == "" || t.Namespace == namespace) {
						for k, v := range r.Variables {
							variables[k] = v
						}
						match = true
						break
					}
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
}

func parseApiVersion(apiVersion string) (string, string) {
	splitted := strings.SplitN(apiVersion, "/", 2)
	if len(splitted) == 2 {
		return splitted[0], splitted[1]
	}
	return "", splitted[0]
}
