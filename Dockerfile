# syntax=docker/dockerfile:1

# --- build stage ---
FROM golang:1.26 AS build
WORKDIR /src

# Cache dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Build the statically-linked binary. Templates and SQL migrations are embedded
# via //go:embed, so the resulting binary is self-contained.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/mayigoo .

# --- runtime stage ---
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/mayigoo /mayigoo
EXPOSE 5000
ENTRYPOINT ["/mayigoo"]
