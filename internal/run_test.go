package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestRun(t *testing.T) {
	assert.NoError(t, Run())
}

func TestReplace(t *testing.T) {
	testCases := []struct {
		config Config
		input  string
		output string
		label  string
		error  error
	}{
		{
			config: Config{},
			input: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: "sub.${DOMAIN:-domain.com}"
`,
			output: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: "sub.${DOMAIN:-domain.com}"
`,
			label: "no-config",
		},
		{
			config: Config{
				Replacements: []Replacement{
					{
						Targets: []Target{
							{
								Kind: "Ingress",
							},
						},
					},
				},
			},
			input: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: "sub.${DOMAIN:-domain.com}"
`,
			output: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: sub.domain.com
`,
			label: "no-variable",
		},
		{
			config: Config{
				Replacements: []Replacement{
					{
						Targets: []Target{
							{
								Kind: "Ingress",
							},
						},
						Variables: map[string]string{
							"DOMAIN": "mydomain.com",
						},
					},
				},
			},
			input: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: "sub.${DOMAIN:-domain.com}"
`,
			output: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: sub.mydomain.com
`,
			label: "with-variable",
		},
		{
			config: Config{
				Replacements: []Replacement{
					{
						Targets: []Target{
							{
								Kind: "Service",
							},
						},
						Variables: map[string]string{
							"PORT": "8080",
						},
					},
				},
			},
			input: `apiVersion: v1
kind: Service
metadata:
  name: service
spec:
  ports:
  - name: http
    port: ${PORT:number}
`,
			output: `apiVersion: v1
kind: Service
metadata:
  name: service
spec:
  ports:
  - name: http
    port: 8080
`,
			label: "number-variable",
		},
		{
			config: Config{
				Replacements: []Replacement{
					{
						Targets: []Target{
							{
								Kind: "Ingress",
							},
						},
						Variables: map[string]string{
							"DOMAIN1": "mydomain1.com",
						},
					},
					{
						Targets: []Target{
							{
								Name: "ingress",
							},
						},
						Variables: map[string]string{
							"DOMAIN2": "mydomain2.com",
						},
					},
				},
			},
			input: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: "sub.${DOMAIN1:-domain.com}"
  - host: "sub.${DOMAIN2:-domain.com}"
`,
			output: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: sub.mydomain1.com
  - host: sub.mydomain2.com
`,
			label: "multiple-replacements",
		},
		{
			config: Config{
				Replacements: []Replacement{
					{
						Variables: map[string]string{
							"DOMAIN": "mydomain.com",
						},
					},
				},
			},
			input: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: "sub.${DOMAIN}"
`,
			output: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: sub.mydomain.com
`,
			label: "empty-targets",
		},
		{
			config: Config{
				Replacements: []Replacement{
					{
						Variables: map[string]string{},
					},
				},
			},
			input: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: "sub.${DOMAIN}"
`,
			output: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: "sub.${DOMAIN}"
`,
			label: "missing",
			error: fmt.Errorf("variable DOMAIN is missing"),
		},
		{
			config: Config{
				Prefix: "PREFIX_",
				Replacements: []Replacement{
					{
						Variables: map[string]string{
							"DOMAIN": "mydomain.com",
						},
					},
				},
			},
			input: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: "sub.${UNKNOWN}"
  - host: "sub.${PREFIX_DOMAIN}"
`,
			output: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
  - host: sub.${UNKNOWN}
  - host: sub.mydomain.com
`,
			label: "prefixed",
		},
	}

	for _, testCase := range testCases {
		inputYaml := yaml.Node{}
		if assert.NoError(t, yaml.Unmarshal([]byte(testCase.input), &inputYaml), testCase.label) {
			outputNodes, err := replace(&testCase.config)([]*yaml.RNode{yaml.NewRNode(&inputYaml)})
			if testCase.error != nil {
				assert.EqualError(t, err, testCase.error.Error(), testCase.label, testCase.label)
			} else if assert.NoError(t, err, testCase.label) {
				if assert.Equal(t, 1, len(outputNodes), testCase.label) {
					outputBytes, err := yaml.Marshal(outputNodes[0].YNode())
					if assert.NoError(t, err, testCase.label) {
						output := string(outputBytes)
						assert.Equal(t, testCase.output, output, testCase.label)
					}
				}
			}
		}
	}
}
