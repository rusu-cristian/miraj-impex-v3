# Impex Line API image (SPEC §11.13, plan D2 and D7). Build context: backend/.
#   docker build -t impex-api .
# M0 builds api and create-owner only; the migrate-history binary is added at M1 (D7).
# The binaries embed the migrations and tzdata, so the runtime image needs nothing else.

# Same Go as go.mod's toolchain line (go1.27.4). GOTOOLCHAIN=local: never download another toolchain here.
FROM golang:1.27.4-bookworm AS build
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && go build -trimpath -ldflags="-s -w" -o /out/create-owner ./cmd/create-owner

# Static, no shell. 65532:65532 is the image's nonroot:nonroot (SPEC §11.13), numeric so runtimes can verify non-root.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ /usr/local/bin/
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
CMD ["serve"]
