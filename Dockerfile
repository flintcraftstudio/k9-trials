FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build CSS
RUN apk add --no-cache curl && \
    curl -sL https://github.com/nicolo-ribaudo/tailwindcss/releases/download/v3.4.17/tailwindcss-linux-x64 -o /usr/local/bin/tailwindcss && \
    chmod +x /usr/local/bin/tailwindcss && \
    tailwindcss -i tailwind/input.css -o web/static/css/site.css --minify

# Generate templ and build Go binary
RUN go install github.com/a-h/templ/cmd/templ@latest && \
    templ generate
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/server /bin/server
COPY --from=build /src/web /web

EXPOSE 8080
ENTRYPOINT ["/bin/server"]
