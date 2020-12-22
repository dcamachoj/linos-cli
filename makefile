.PHONY: install

install:
	@go install .

build-win:
	@GOOS=windows GOARCH=amd64 go build -o dist/linos-cli.exe main.go

build-mac:
	@GOOS=darwin GOARCH=amd64 go build -o dist/linos-cli.mac main.go

build-linux:
	@GOOS=linux GOARCH=amd64 go build -o dist/linos-cli.linux main.go

build: build-win build-mac build-linux