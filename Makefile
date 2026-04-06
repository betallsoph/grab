.PHONY: run build swagger

run: swagger
	go run cmd/api/main.go

build: swagger
	go build -o bin/grab cmd/api/main.go

swagger:
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
