# Build stage
FROM golang:1.22.1 AS builder

WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -o /go/bin/ctfd-setup cmd/ctfd-setup/main.go



# Prod stage
FROM scratch
COPY --from=builder /go/bin/ctfd-setup /ctfd-setup
ENTRYPOINT [ "/ctfd-setup" ]
