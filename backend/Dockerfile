FROM golang:1.26.3-alpine AS source
WORKDIR /src
ENV GOTOOLCHAIN=local \
    GOTMPDIR=/src/.cache/work \
    TMPDIR=/src/.cache/work \
    GOCACHE=/src/.cache/build
RUN mkdir -p .cache/work .cache/build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM source AS test
RUN --mount=type=cache,target=/src/.cache/build \
    CGO_ENABLED=0 go test ./... && go vet ./...

FROM test AS build
RUN --mount=type=cache,target=/src/.cache/build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /src/rpg-backend ./cmd/

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /src/rpg-backend /rpg-backend
EXPOSE 8080
ENTRYPOINT ["/rpg-backend"]
