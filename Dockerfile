FROM golang:1.23.2@sha256:ad5c126b5cf501a8caef751a243bb717ec204ab1aa56dc41dc11be089fafcb4f AS builder

WORKDIR /go/src
COPY go.mod go.sum .
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
ARG VERSION="dev"
ARG COMMIT
ARG COMMIT_DATE
RUN go build -cover \
    -ldflags="-s -w -X 'main.Version="$VERSION"' -X 'main.Commit="$COMMIT"' -X 'main.CommitDate="$COMMIT_DATE"'" \
    -o /go/bin/ctfd-setup \
    cmd/ctfd-setup/main.go



FROM scratch
COPY --from=builder /go/bin/ctfd-setup /ctfd-setup
ENTRYPOINT [ "/ctfd-setup" ]
