all: generate build

.PHONY: docker-image
docker-image:
	docker build -t ereslibre/cluster-api-provider-proxmox:latest .

.PHONY: generate
generate:
	go generate ./pkg/... ./cmd/...

.PHONY: build
build:
	mkdir -p ./bin
	go build -o ./bin/manager ./cmd/manager
	go build -o ./bin/clusterctl ./cmd/clusterctl

.PHONY: install
install:
	CGO_ENABLED=0 go install -ldflags '-extldflags "-static"' github.com/ereslibre/cluster-api-provider-proxmox/cmd/manager
	CGO_ENABLED=0 go install -ldflags '-extldflags "-static"' github.com/ereslibre/cluster-api-provider-proxmox/cmd/clusterctl

.PHONY: deploy
deploy: manifests
	cat provider-components.yaml | kubectl apply -f -

.PHONY: manifests
manifests:
	kustomize build vendor/sigs.k8s.io/cluster-api/config/default/ > cmd/clusterctl/examples/proxmox/out/provider-components.yaml
	echo "---" >> cmd/clusterctl/examples/proxmox/out/provider-components.yaml
	$(MAKE) -C cmd/clusterctl/examples/proxmox examples
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go crd
	kustomize build config/default >> cmd/clusterctl/examples/proxmox/out/provider-components.yaml
