-- Script de inicialización de base de datos PostgreSQL
-- Para GridFlow-Dynamics

-- Crear tabla de cuadrillas
CREATE TABLE IF NOT EXISTS cuadrillas (
    id SERIAL PRIMARY KEY,
    grupo_trabajo VARCHAR(255) NOT NULL,
    nombre_empleado VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    latitud DECIMAL(9,6) NOT NULL,
    longitud DECIMAL(9,6) NOT NULL,
    codigo_odt VARCHAR(255) NOT NULL,
    estado VARCHAR(50) NOT NULL CHECK (estado IN ('en_ruta', 'trabajando', 'en_pausa', 'finalizado')),
    porcentaje_progreso INT NOT NULL CHECK (porcentaje_progreso >= 0 AND porcentaje_progreso <= 100),
    nivel_bateria INT NOT NULL CHECK (nivel_bateria >= 0 AND nivel_bateria <= 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Crear índices para mejorar rendimiento de consultas
CREATE INDEX idx_cuadrillas_grupo_trabajo ON cuadrillas(grupo_trabajo);
CREATE INDEX idx_cuadrillas_codigo_odt ON cuadrillas(codigo_odt);
CREATE INDEX idx_cuadrillas_timestamp ON cuadrillas(timestamp);
CREATE INDEX idx_cuadrillas_estado ON cuadrillas(estado);

-- Comentarios en tabla y columnas
COMMENT ON TABLE cuadrillas IS 'Tabla de mensajes de inventario de cuadrillas';
COMMENT ON COLUMN cuadrillas.grupo_trabajo IS 'Identificador del grupo de trabajo';
COMMENT ON COLUMN cuadrillas.nombre_empleado IS 'Nombre del empleado';
COMMENT ON COLUMN cuadrillas.timestamp IS 'Marca de tiempo del mensaje';
COMMENT ON COLUMN cuadrillas.latitud IS 'Latitud de la ubicación';
COMMENT ON COLUMN cuadrillas.longitud IS 'Longitud de la ubicación';
COMMENT ON COLUMN cuadrillas.codigo_odt IS 'Código de la orden de trabajo';
COMMENT ON COLUMN cuadrillas.estado IS 'Estado de la cuadrilla';
COMMENT ON COLUMN cuadrillas.porcentaje_progreso IS 'Porcentaje de progreso del trabajo';
COMMENT ON COLUMN cuadrillas.nivel_bateria IS 'Nivel de batería del dispositivo';
