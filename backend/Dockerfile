# Impex Line API image (SPEC §11.13, plan D2 and D7). Build context: backend/.
#   docker build -t impex-api .
# Three binaries: api (serve, migrate), create-owner, and migrate-history (one-off history load, M1, D7).
# The binaries embed the migrations and tzdata, so the runtime image needs nothing else; migrate-history reads the
# workbook extracts from a mounted directory (--extracts), they are not in the image.

# Same Go as go.mod's toolchain line (go1.27.4). GOTOOLCHAIN=local: never download another toolchain here.
# Base images are pinned tag@digest (multi-arch index digest; the digest wins). To bump: NOTES-verify.md, "Base image digests".
FROM golang:1.27.4-bookworm@sha256:ce59b41415e2524ae99926fe1f37ce48ba8c194dcbd2a56a65436cdcefa2bd34 AS build
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && go build -trimpath -ldflags="-s -w" -o /out/create-owner ./cmd/create-owner \
 && go build -trimpath -ldflags="-s -w" -o /out/migrate-history ./cmd/migrate-history

# Static, no shell. 65532:65532 is the image's nonroot:nonroot (SPEC §11.13), numeric so runtimes can verify non-root.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /out/ /usr/local/bin/
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
CMD ["serve"]
