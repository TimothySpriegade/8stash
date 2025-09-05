COVERPROFILE ?= coverage.out
PACKAGES ?= ./...

test:
	go test -coverprofile="$(COVERPROFILE)" -covermode=atomic $(PACKAGES)

install-local:
	(cd ./scripts && ./build_install_locally.sh)

