PROJECT_DIR = $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

GOMAXPROCS = 4
GOVERSION  = 1.10

PROJECT = "github.com/sethvargo/vault-token-helper-gcp-kms"
NAME    = $(shell go run version/cmd/main.go name)
VERSION = $(shell go run version/cmd/main.go version)
COMMIT  = $(shell git rev-parse --short HEAD)

LDFLAGS = \
	-s \
	-w \
	-X ${PROJECT}/version.Name=${NAME} \
	-X ${PROJECT}/version.GitCommit=${COMMIT}

# XC_* are the platforms for cross-compiling. Customize these values to suit
# your needs.
XC_OS      = darwin linux windows
XC_ARCH    = 386 amd64
XC_EXCLUDE =

# deps updates the project deps using golang/dep
deps:
	@dep ensure -update
.PHONY: deps

# dev builds and installs the plugin for local development
dev:
	@env \
		CGO_ENABLED=0 \
		GOMAXPROCS="${GOMAXPROCS}" \
		go install \
			-ldflags="${LDFLAGS}"
.PHONY: dev

# docker builds the docker container
docker:
	@docker build \
		--tag "sethvargo/${NAME}" \
		--tag "sethvargo/${NAME}:${VERSION}" \
		. && \
	@docker push "sethvargo/${NAME}" && \
	@docker push "sethvargo/${NAME}:${VERSION}"
.PHONY: docker

# pkg creates a single directory of all the files and signs them with gpg
pkg:
	@rm -rf "${PROJECT_DIR}/pkg/dist"
	@mkdir -p "${PROJECT_DIR}/pkg/dist"
	@for OS in $(XC_OS); do \
		for ARCH in $(XC_ARCH); do \
			gpg --detach-sign -o "${PROJECT_DIR}/pkg/$${OS}_$${ARCH}/${NAME}.sig" "${PROJECT_DIR}/pkg/$${OS}_$${ARCH}/${NAME}" ; \
			cp "${PROJECT_DIR}/pkg/$${OS}_$${ARCH}/${NAME}.sig" "${PROJECT_DIR}/pkg/dist/${NAME}_$${OS}_$${ARCH}.sig" ; \
			cp "${PROJECT_DIR}/pkg/$${OS}_$${ARCH}/${NAME}" "${PROJECT_DIR}/pkg/dist/${NAME}_$${OS}_$${ARCH}" ; \
		done \
	done
.PHONY: pkg

# test runs the tests
test:
	@go test -timeout=30s -parallel=20 ./...
.PHONY: test

# xc compiles all the binaries using containers as an isolation layer
xc:
	@rm -rf "${PROJECT_DIR}/pkg"
	@for OS in $(XC_OS); do \
		for ARCH in $(XC_ARCH); do \
			if [[ "${XC_EXCLUDE}" != *"$${OS}/$${ARCH}"* ]]; then \
				echo "--> $${OS}/$${ARCH}" ; \
				docker run \
					--interactive \
					--tty \
					--rm \
					--dns="1.1.1.1" \
					--volume="${PROJECT_DIR}:/go/src/${PROJECT}" \
					--workdir="/go/src/${PROJECT}" \
					golang:${GOVERSION}-alpine \
			 			env \
							CGO_ENABLED="0" \
							GOOS="$${OS}" \
							GOARCH="$${ARCH}" \
							GOMAXPROCS="${GOMAXPROCS}" \
							go build \
								-a \
								-o "pkg/$${OS}_$${ARCH}/${NAME}" \
								-ldflags "${LDFLAGS}" ; \
			fi ; \
		done \
	done
.PHONY: xc
