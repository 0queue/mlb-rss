FROM golang:1.18.1 AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
# not very good for the docker build
# cache but it sure saves time
COPY . ./
RUN CGO_ENABLED=0 go build -o /mlb-rss

FROM scratch
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=builder /mlb-rss /mlb-rss
ENTRYPOINT ["/mlb-rss"]
