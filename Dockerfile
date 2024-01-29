# syntax=docker/dockerfile:1

FROM golang:1.21 AS builder

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid 65532 \
  small-user

WORKDIR /app
RUN export GOPATH=$GOPATH:$(pwd)
COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod verify

COPY . /app/
ARG GOARCH=amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH go build -o /main main.go

FROM gcr.io/distroless/static-debian12

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /main .

USER small-user:small-user
EXPOSE 2112

CMD [ "./main" ]
