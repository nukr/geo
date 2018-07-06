clean:
	@rm main
build:
	@GOOS=linux GOARCH=amd64 go build cmd/rest/main.go
build-docker:
	@docker build -t asia.gcr.io/instant-matter-785/address:0.0.7 .
build-all:
	@make build && make build-docker && make clean
