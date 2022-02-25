# kustomize-variable-injector

Plugin for [kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/) to allow to expand placeholders with values. The initial use case is to make domain names in ingresses configurable.

## Syntax

* `host: sub.${DOMAIN}` replace with value `DOMAIN` or fail if value is missing
* `host: sub.${DOMAIN:-fallback.com}` replace with value `DOMAIN` or fall back to "fallback.com" if missing
* `port: ${REDIS_PORT:number:-6379}` replace with value `REDIS_PORT`, inject as number (not as string) and fall back to 6379 if missing

## Usage

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
spec:
  rules:
    - host: "sub.${DOMAIN:-domain.com}"

# variable-injector.yaml
apiVersion: airfocusio.github.com/v1alpha1
kind: VariableReplacer
metadata:
  name: variable-injector
  annotations:
    config.kubernetes.io/function: |
      container:
        image: ghcr.io/airfocusio/kustomize-variable-injector:latest
replacements:
  - targets:
      - group: networking.k8s.io
        version: v1
        kind: Ingress
        name: ingress
    variables:
      DOMAIN: mydomain.com

# kustomization.yaml
resources:
  - secret.yaml
transformers:
  - variable-injector.yaml
```

## Caveats

Plugins are still in alpha. For this to work, you need to provide the `--enable-alpha-plugins` flag (i.e. `kustomize build --enable-alpha-plugins`).
