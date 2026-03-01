FROM golang:1.26.0@sha256:9edf71320ef8a791c4c33ec79f90496d641f306a91fb112d3d060d5c1cee4e20 AS builder

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
