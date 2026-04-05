docker-up-build:
	docker compose -f docker-compose-dev.yml up --build

docker-up:
	docker compose -f docker-compose-dev.yml up

docker-down:
	docker compose -f docker-compose-dev.yml down

docker-down-rm-vol:
	docker compose -f docker-compose-dev.yml down -v

swagger-docs:
	cd services/user-service && swag init -g cmd/main.go -d ./,../../common
	cd services/banking-service && swag init -g cmd/main.go -d ./,../../common
	cd services/trading-service && swag init -g cmd/main.go -d ./,../../common

proto:
	protoc --proto_path=. --go_out=. --go-grpc_out=. common/proto/*.proto

test:
	go test ./common/... ./services/user-service/... ./services/banking-service/... ./services/trading-service/...

test-integration:
	go test -tags=integration ./common/... ./services/user-service/... ./services/banking-service/...

coverage-integration-profile:
	mkdir -p .tmp-coverage
	go test -tags=integration -covermode=count -coverpkg=$$(go list ./common/... ./services/user-service/... ./services/banking-service/... | grep -v '/internal/integration_test$$' | paste -sd, -) -coverprofile=.tmp-coverage/coverage.integration.out ./common/... ./services/user-service/... ./services/banking-service/...

coverage-integration: coverage-integration-profile
	go tool cover -func=.tmp-coverage/coverage.integration.out | tail -n 1

coverage-html: coverage-integration-profile
	go tool cover -html=.tmp-coverage/coverage.integration.out
