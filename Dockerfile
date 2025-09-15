FROM golang:1.25 AS build
WORKDIR /app
COPY go.mod .
COPY . ./
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go build -o /bin/server ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /bin/server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
