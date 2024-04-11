run: build
	@./go_redis

build:
	@go build -o ./go_redis .
