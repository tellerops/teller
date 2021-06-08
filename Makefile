setup-mac:
	brew install golangci-lint
	go install github.com/golang/mock/mockgen@v1.5.0
mocks:
	mockgen -source pkg/providers/aws_secretsmanager.go -destination pkg/providers/mock_providers/aws_secretsmanager_mock.go
	mockgen -source pkg/providers/aws_ssm.go -destination pkg/providers/mock_providers/aws_ssm_mock.go
	mockgen -source pkg/providers/consul.go -destination pkg/providers/mock_providers/consul_mock.go
	mockgen -source pkg/providers/dotenv.go -destination pkg/providers/mock_providers/dotenv_mock.go
	mockgen -source pkg/providers/doppler.go -destination pkg/providers/mock_providers/doppler_mock.go
	mockgen -source pkg/providers/etcd.go -destination pkg/providers/mock_providers/etcd_mock.go
	mockgen -source pkg/providers/google_secretmanager.go -destination pkg/providers/mock_providers/google_secretmanager_mock.go
	mockgen -source pkg/providers/hashicorp_vault.go -destination pkg/providers/mock_providers/hashicorp_vault_mock.go
	mockgen -source pkg/providers/heroku.go -destination pkg/providers/mock_providers/heroku_mock.go
	mockgen -source pkg/providers/vercel.go -destination pkg/providers/mock_providers/vercel_mock.go
readme:
	yarn readme
lint:
	golangci-lint run
test:
	go test -v ./pkg/... -cover

integration:
	go test -v ./pkg/integration_test -cover -tags=integration

integration_api:
	go test -v ./pkg/integration_test -cover -tags="integration_api integration"

deps:
	go mod tidy && go mod vendor

release:
	goreleaser --rm-dist

coverage:
	go test ./pkg/... -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: deps setup-mac release readme lint mocks coverage
