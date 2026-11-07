# Backend Go (CU-01..CU-10). Imagen final mínima, sin shell ni compilador.
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/opl ./cmd/opl \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/opl /out/seed /usr/local/bin/
EXPOSE 8080
USER nonroot
ENTRYPOINT ["/usr/local/bin/opl"]
