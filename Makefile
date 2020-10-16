IMAGE=container.trusch.io/caddy-extauth/caddy:latest
DUMMY_AUTH_IMAGE=container.trusch.io/caddy-extauth/dummy-auth:latest
HTTP_LOGGER_IMAGE=container.trusch.io/caddy-extauth/http-logger:latest
BASE_IMAGE=gcr.io/distroless/base-debian10:latest
BUILD_IMAGE=container.trusch.io/caddy-extauth/builder
BUILD_BASE_IMAGE=golang:1.15

all: image dummy-auth-image http-logger-image

bin/caddy: extauth.go .build-image
	mkdir -p bin
	podman run \
		--rm \
		-v ./:/app \
		-w /app \
		-e GOOS=${GOOS} \
		-e GOARCH=${GOARCH} \
		-e GOARM=${GOARM} \
		-v go-build-cache:/root/.cache/go-build \
		-v go-mod-cache:/go/pkg/mod $(BUILD_IMAGE) bash -c \
			"xcaddy build master --with github.com/trusch/caddy-extauth/pkg/extauth=/app && mv caddy bin/caddy"

bin/http-logger: cmd/http-logger/main.go
	mkdir -p bin
	podman run \
		--rm \
		-v ./:/app \
		-w /app \
		-e GOOS=${GOOS} \
		-e GOARCH=${GOARCH} \
		-e GOARM=${GOARM} \
		-v go-build-cache:/root/.cache/go-build \
		-v go-mod-cache:/go/pkg/mod $(BUILD_IMAGE) \
			go build -o $@ ./cmd/http-logger

bin/dummy-auth: cmd/dummy-auth/main.go
	mkdir -p bin
	podman run \
		--rm \
		-v ./:/app \
		-w /app \
		-e GOOS=${GOOS} \
		-e GOARCH=${GOARCH} \
		-e GOARM=${GOARM} \
		-v go-build-cache:/root/.cache/go-build \
		-v go-mod-cache:/go/pkg/mod $(BUILD_IMAGE) \
			go build -o $@ ./cmd/dummy-auth

build-image: .build-image
.build-image:
	$(eval ID=$(shell buildah from $(BUILD_BASE_IMAGE)))
	buildah run $(ID) go get -u github.com/caddyserver/xcaddy/cmd/xcaddy
	buildah commit $(ID) $(BUILD_IMAGE)
	buildah rm $(ID)
	touch .build-image

image: .image
.image: bin/caddy
	$(eval ID=$(shell buildah from $(BASE_IMAGE)))
	buildah copy $(ID) bin/caddy /bin/
	buildah commit $(ID) $(IMAGE)
	buildah rm $(ID)
	touch .image

dummy-auth-image: .dummy-auth-image
.dummy-auth-image: bin/dummy-auth
	$(eval ID=$(shell buildah from $(BASE_IMAGE)))
	buildah copy $(ID) bin/dummy-auth /bin/
	buildah config --cmd dummy-auth $(ID)
	buildah commit $(ID) $(DUMMY_AUTH_IMAGE)
	buildah rm $(ID)
	touch .dummy-auth-image

http-logger-image: .http-logger-image
.http-logger-image: bin/http-logger
	$(eval ID=$(shell buildah from $(BASE_IMAGE)))
	buildah copy $(ID) bin/http-logger /bin/
	buildah config --cmd http-logger $(ID)
	buildah commit $(ID) $(HTTP_LOGGER_IMAGE)
	buildah rm $(ID)
	touch .http-logger-image

clean:
	-rm -r .build-image .image .http-logger-image .dummy-auth-image bin
