# =========================================================
# Presenz — Go Attendance System
# Multi-stage build: compiles a fully static binary, ships
# it in a scratch image with zero OS packages / zero base-
# image CVE surface.
# =========================================================

# -------------------------
# Build Stage
# -------------------------
FROM golang:1.26.5-alpine3.24 AS builder

# Patch whatever Alpine snapshot this Go image ships with,
# at build time — keeps the builder stage reasonably current
# without pinning to a specific Alpine minor version that can
# drift stale between rebuilds.
RUN apk update && apk upgrade --no-cache

WORKDIR /build

# Cache dependency downloads separately from source changes
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 works because modernc.org/sqlite is pure Go —
# no cgo cross-compilation needed, and it's what makes the
# scratch final stage possible (fully static binary).
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/presenz ./cmd/presenz

# -------------------------
# Final Stage
# -------------------------
FROM scratch

WORKDIR /app

# entry.html is go:embed'd into the binary itself (internal/client) —
# nothing to copy for it.
#
# config.json is NOT embedded — read from disk at runtime via
# config.LoadConfig — so it must ship in the image.
COPY --from=builder /build/presenz /app/presenz
COPY config/config.json /app/config/config.json

# db/ and backup/ are created at runtime by the app itself
# (os.MkdirAll in Connect/Export). Mount volumes at `docker run`
# time if you want data to persist outside the container:
#   -v ./db:/app/db -v ./backup:/app/backup

EXPOSE 8080

# No default CLI args baked in — pass --course/--batch/--total
# after the image name at `docker run` time, e.g.:
#   docker run -p 8080:8080 presenz --course CSE --batch B1 --total 60
ENTRYPOINT ["/app/presenz"]