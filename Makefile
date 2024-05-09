# Run go fmt against code
fmt:
	go fmt ./...
	gofmt -s -w .

# Run go vet against code
vet:
	go vet ./...

# Run golangci-lint
lint:
	golangci-lint run

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test: mocks tidy fmt vet
	go test ./... -coverprofile cover.out -timeout 30m

mocks: mockgen
	mockgen -destination pkg/mocks/manager/mock.go sigs.k8s.io/controller-runtime/pkg/manager Manager

mockgen:
ifeq (, $(shell which mockgen))
 $(shell go install go.uber.org/mock/mockgen@v0.4.0)
endif