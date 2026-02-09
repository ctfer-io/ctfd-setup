FROM golang:1.25.4@sha256:698183780de28062f4ef46f82a79ec0ae69d2d22f7b160cf69f71ea8d98bf25d AS builder

WORKDIR /go/src
COPY go.mod go.sum .
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
ARG VERSION="dev"
ARG COMMIT
ARG DATE
RUN go build -cover \
    -ldflags="-s -w -X 'main.version="$VERSION"' -X 'main.commit="$COMMIT"' -X 'main.date="$DATE"' -X 'main.builtBy=docker'" \
    -o /go/bin/ctfd-setup \
    cmd/ctfd-setup/main.go



FROM scratch
COPY --from=builder /go/bin/ctfd-setup /ctfd-setup
ENTRYPOINT [ "/ctfd-setup" ]
