FROM golang:1.26.1-alpine AS build
RUN apk add --no-cache ca-certificates make
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY Makefile ./
RUN make build BINARY=/http_mail

FROM scratch
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /http_mail /app/http_mail
EXPOSE 8000
ENTRYPOINT ["/app/http_mail"]
