NAME := crypto
SOURCE_FILES := $(shell find * -name '*.go')
GIT_COMMIT := $(shell git describe --dirty --always)
VERSION := 1.2

cloud-platform-git-xargs: $(SOURCE_FILES)
	go mod download
	GIT_COMMIT=$$(git rev-list -1 HEAD) && CGO_ENABLED=0 go build  -ldflags -a -o ./cloud-platform-git-xargs ./main.go

test:
	go test -v ./...

fmt:
	gofmt -l -s -w ./
	goimports -l -w ./

release:
	git tag $(VERSION)
	git push origin $(VERSION)
