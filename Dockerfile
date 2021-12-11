FROM alpine:latest as certs
RUN apk add --update --no-cache ca-certificates
COPY kustomize-variable-injector /bin/kustomize-variable-injector
ENTRYPOINT ["/bin/kustomize-variable-injector"]
WORKDIR /workdir
