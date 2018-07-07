GIT_LATEST_TAG = $(shell sh -c 'git tag | tail -n 1 | cut -c 2-')
clean:
	@rm main
build:
	@GOOS=linux GOARCH=amd64 go build cmd/rest/main.go
build-docker:
	@docker build -t asia.gcr.io/instant-matter-785/geo:$(GIT_LATEST_TAG) .
build-all:
	@make build && make build-docker && make clean
