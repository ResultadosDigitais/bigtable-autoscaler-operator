FROM golang:1.13.0-buster AS builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on make

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/bin/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
