# syntax=docker/dockerfile:1

# ---- Build stage: compile static binaries ----
FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck

# ---- Runtime stage: minimal, non-root image ----
FROM gcr.io/distroless/static:nonroot

COPY --from=build /out/server /server
COPY --from=build /out/healthcheck /healthcheck

USER nonroot:nonroot
ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 CMD ["/healthcheck"]

ENTRYPOINT ["/server"]
