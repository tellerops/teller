setup-mac:
	brew install golangci-lint
	go install github.com/golang/mock/mockgen@v1.5.0
mocks:
	mockgen -source pkg/providers/aws_secretsmanager.go -destination pkg/providers/mock_providers/aws_secretsmanager_mock.go
	mockgen -source pkg/providers/aws_ssm.go -destination pkg/providers/mock_providers/aws_ssm_mock.go
	mockgen -source pkg/providers/cloudflare_workers_kv.go -destination pkg/providers/mock_providers/cloudflare_workers_kv_mock.go
	mockgen -source pkg/providers/cloudflare_workers_secrets.go -destination pkg/providers/mock_providers/cloudflare_workers_secrets_mock.go
	mockgen -source pkg/providers/consul.go -destination pkg/providers/mock_providers/consul_mock.go
	mockgen -source pkg/providers/dotenv.go -destination pkg/providers/mock_providers/dotenv_mock.go
	mockgen -source pkg/providers/doppler.go -destination pkg/providers/mock_providers/doppler_mock.go
	mockgen -source pkg/providers/etcd.go -destination pkg/providers/mock_providers/etcd_mock.go
	mockgen -source pkg/providers/google_secretmanager.go -destination pkg/providers/mock_providers/google_secretmanager_mock.go
	mockgen -source pkg/providers/hashicorp_vault.go -destination pkg/providers/mock_providers/hashicorp_vault_mock.go
	mockgen -source pkg/providers/heroku.go -destination pkg/providers/mock_providers/heroku_mock.go
	mockgen -source pkg/providers/vercel.go -destination pkg/providers/mock_providers/vercel_mock.go
	mockgen -source pkg/providers/onepassword.go -destination pkg/providers/mock_providers/onepassword_mock.go
	mockgen -source pkg/providers/gopass.go -destination pkg/providers/mock_providers/gopass_mock.go
	mockgen -source pkg/providers/github.go -destination pkg/providers/mock_providers/github_mock.go
	mockgen -source pkg/providers/azure_keyvault.go -destination pkg/providers/mock_providers/azure_keyvault_mock.go
	mockgen -source pkg/providers/keeper_secretsmanager.go -destination pkg/providers/mock_providers/keeper_secretsmanager_mock.go
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

build:
	go build -ldflags "-s -w -X main.version=0.0.0 -X main.commit=0000000000000000000000000000000000000000 -X main.date=2022-01-01"

e2e: build
	BINARY_PATH="$(shell pwd)/teller" go test -v ./e2e
	
coverage:
	go test ./pkg/... -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: deps setup-mac release readme lint mocks coverage
