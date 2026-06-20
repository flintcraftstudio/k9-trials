FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build CSS
RUN apk add --no-cache curl && \
    curl -sL https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.17/tailwindcss-linux-x64 -o /usr/local/bin/tailwindcss && \
    chmod +x /usr/local/bin/tailwindcss && \
    tailwindcss -i tailwind/input.css -o web/static/css/site.css --minify

# Generate templ + sqlc code (both output to gitignored dirs), then build.
# sqlc is pinned to match local generation (mage generate) so the generated
# internal/db package matches what the handlers were developed against.
RUN go install github.com/a-h/templ/cmd/templ@latest && \
    templ generate
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 && \
    sqlc generate
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/server /bin/server
COPY --from=build /src/web /web
# goose.Up reads migrations from disk at startup (SetBaseFS(nil)), relative to
# the working dir, so the migrations dir must ship in the runtime image.
COPY --from=build /src/migrations /migrations

EXPOSE 8080
ENTRYPOINT ["/bin/server"]
