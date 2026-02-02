# Multi-stage build para imagen mínima
# Etapa 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Instalar dependencias de build
RUN apk add --no-cache git ca-certificates tzdata

# Copiar archivos de módulos Go primero (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar binario estático optimizado
# CGO_ENABLED=0: binario estático sin dependencias C
# -ldflags="-w -s": eliminar símbolos de debug y reduce tamaño
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/gridflow-server \
    ./cmd/server

# Etapa 2: Runtime - Imagen mínima
FROM scratch

# Copiar certificados CA para HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copiar zona horaria
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copiar binario compilado
COPY --from=builder /build/gridflow-server /gridflow-server

# Exponer puerto de la API
EXPOSE 8080

# Usuario no-root (opcional, scratch no tiene usuarios)
# Ejecutar aplicación
ENTRYPOINT ["/gridflow-server"]
