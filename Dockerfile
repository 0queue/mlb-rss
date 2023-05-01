FROM docker.io/library/golang:1.20.3 AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
# not very good for the docker build
# cache but it sure saves time
COPY . ./
RUN CGO_ENABLED=0 go build -o /mlb-rss cmd/mlb-rss/main.go

FROM gcr.io/distroless/static-debian11
COPY --from=builder /mlb-rss /mlb-rss
ENV JSON_LOG=true
ENTRYPOINT ["/mlb-rss"]
