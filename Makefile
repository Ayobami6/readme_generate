build:
	@go build -o ./bin/readme

run: build
	@./bin/readme