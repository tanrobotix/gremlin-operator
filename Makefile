VERSION_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GO_FLAGS ?= GOOS=linux GOARCH=amd64 CGO_ENABLED=0
KUBERNETES_CONFIG ?= "$(HOME)/.kube/config"
WATCH_NAMESPACE ?= default
BIN_DIR ?= "build/_output/bin"
IMPORT_LOG=import.log
FMT_LOG=fmt.log

OPERATOR_NAME ?= gremlin-operator
NAMESPACE ?= "$(USER)"
BUILD_IMAGE ?= "$(NAMESPACE)/$(OPERATOR_NAME):latest"
OUTPUT_BINARY ?= "$(BIN_DIR)/$(OPERATOR_NAME)"
VERSION_PKG ?= "github.com/Kubedex/gremlin-operator/pkg/version"
GREMLIN_VERSION ?= "$(shell grep -v '\#' gremlin.version)"
OPERATOR_VERSION ?= "$(shell git describe --tags)"

LD_FLAGS ?= "-X $(VERSION_PKG).version=$(OPERATOR_VERSION) -X $(VERSION_PKG).buildDate=$(VERSION_DATE) -X $(VERSION_PKG).defaultGremlin=$(GREMLIN_VERSION)"
PACKAGES := $(shell go list ./cmd/... ./pkg/...)

.DEFAULT_GOAL := build

.PHONY: check
check:
	@echo Checking...
	@go fmt $(PACKAGES) > $(FMT_LOG)
	@[ ! -s "$(FMT_LOG)" -a ! -s "$(IMPORT_LOG)" ] || (echo "Go fmt, license check, or import ordering failures, run 'make format'" | cat - $(FMT_LOG) $(IMPORT_LOG) && false)

.PHONY: ensure-generate-is-noop
ensure-generate-is-noop: generate
	@git diff -s --exit-code pkg/apis/gremlin/v1alpha1/zz_generated.deepcopy.go || (echo "Build failed: a model has been changed but the deep copy functions aren't up to date. Run 'make generate' and update your PR." && exit 1)

.PHONY: format
format:
	@echo Formatting code...
	@go fmt $(PACKAGES)

.PHONY: lint
lint:
	@echo Linting...
	@golint $(PACKAGES)
	@gosec -quiet -exclude=G104,G101 $(PACKAGES) 2>/dev/null

.PHONY: build
build: format
	@echo Building...
	@operator-sdk build kubedex/gremlin-operator:build

.PHONY: docker
docker:
	@docker tag kubedex/gremlin-operator:build "$(BUILD_IMAGE)"

.PHONY: push
push:
	@echo Pushing image $(BUILD_IMAGE)...
	@docker push $(BUILD_IMAGE) > /dev/null

.PHONY: unit-tests
unit-tests:
	@echo Running unit tests...
	@go test $(PACKAGES) -cover -coverprofile=cover.out

.PHONY: e2e-tests
e2e-tests: cassandra es crd build docker push
	@mkdir -p deploy/test
	@echo Running end-to-end tests...

	@cp test/role_binding.yaml deploy/test/namespace-manifests.yaml
	@echo "---" >> deploy/test/namespace-manifests.yaml

	@cat test/role.yaml >> deploy/test/namespace-manifests.yaml
	@echo "---" >> deploy/test/namespace-manifests.yaml

	@cat test/service_account.yaml >> deploy/test/namespace-manifests.yaml
	@echo "---" >> deploy/test/namespace-manifests.yaml

	@cat test/operator.yaml | sed "s~image: Kubedex\/gremlin-operator\:.*~image: $(BUILD_IMAGE)~gi" >> deploy/test/namespace-manifests.yaml
	@go test ./test/e2e/... -kubeconfig $(KUBERNETES_CONFIG) -namespacedMan ../../deploy/test/namespace-manifests.yaml -globalMan ../../deploy/crds/gremlin_v1alpha1_gremlin_crd.yaml -root .

.PHONY: run
run: crd
	@bash -c 'trap "exit 0" INT; OPERATOR_NAME=${OPERATOR_NAME} KUBERNETES_CONFIG=${KUBERNETES_CONFIG} WATCH_NAMESPACE=${WATCH_NAMESPACE} go run -ldflags ${LD_FLAGS} main.go start'

.PHONY: clean
clean:
	@kubectl delete -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.18.0/deploy/mandatory.yaml || true

.PHONY: crd
crd:
	@kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_crd.yaml 2>&1 | grep -v "already exists" || true

.PHONY: ingress
ingress:
	# see https://kubernetes.github.io/ingress-nginx/deploy/#verify-installation
	@kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.18.0/deploy/mandatory.yaml
	@minikube addons enable ingress

.PHONY: generate
generate:
	@operator-sdk generate k8s

.PHONY: test
test: unit-tests e2e-tests

.PHONY: all
all: check format lint build test

.PHONY: ci
ci: ensure-generate-is-noop check format lint build unit-tests