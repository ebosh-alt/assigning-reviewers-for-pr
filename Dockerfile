FROM golang:1.23-bullseye AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ENV CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOFLAGS='-trimpath'
RUN go build -o /app/assigning-reviewers-for-pr ./cmd

FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=builder /app/assigning-reviewers-for-pr /app/assigning-reviewers-for-pr
COPY --from=builder /app/config /app/config
COPY --from=builder /app/db/migrations /app/db/migrations

EXPOSE 8080

ENTRYPOINT ["/app/assigning-reviewers-for-pr"]
