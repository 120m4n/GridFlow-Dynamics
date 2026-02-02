package domain

import (
	"fmt"
	"time"
)

// EstadoCuadrilla representa el estado de una cuadrilla durante el seguimiento.
type EstadoCuadrilla string

const (
	EstadoEnRuta     EstadoCuadrilla = "en_ruta"
	EstadoTrabajando EstadoCuadrilla = "trabajando"
	EstadoEnPausa    EstadoCuadrilla = "en_pausa"
	EstadoFinalizado EstadoCuadrilla = "finalizado"
)

// Coordenadas representa los datos de ubicación GPS.
type Coordenadas struct {
	Latitud  float64 `json:"latitud"`
	Longitud float64 `json:"longitud"`
}

// MensajeInventarioCuadrilla representa el payload JSON de la app móvil según especificación.
type MensajeInventarioCuadrilla struct {
	GrupoTrabajo        string      `json:"grupoTrabajo"`
	NombreEmpleado      string      `json:"nombreEmpleado"`
	Timestamp           time.Time   `json:"timestamp"`
	Coordenadas         Coordenadas `json:"coordenadas"`
	CodigoODT           string      `json:"codigoODT"`
	Estado              string      `json:"estado"`
	PorcentajeProgreso  int         `json:"procentajeProgreso"`
	NivelBateria        int         `json:"nivelBateria"`
}

// Validar valida todos los campos del mensaje de inventario de cuadrilla.
func (m *MensajeInventarioCuadrilla) Validar() error {
	// Validar grupoTrabajo - cadena no vacía
	if m.GrupoTrabajo == "" {
		return fmt.Errorf("grupoTrabajo es requerido y no puede estar vacío")
	}

	// Validar nombreEmpleado - cadena no vacía
	if m.NombreEmpleado == "" {
		return fmt.Errorf("nombreEmpleado es requerido y no puede estar vacío")
	}

	// Validar codigoODT - cadena no vacía
	if m.CodigoODT == "" {
		return fmt.Errorf("codigoODT es requerido y no puede estar vacío")
	}

	// Validar timestamp - ISO8601 válido
	if m.Timestamp.IsZero() {
		return fmt.Errorf("timestamp es requerido y debe ser una fecha válida en formato ISO8601")
	}

	// Validar coordenadas.latitud: -90 a 90
	if m.Coordenadas.Latitud < -90 || m.Coordenadas.Latitud > 90 {
		return fmt.Errorf("coordenadas.latitud debe estar entre -90 y 90, recibido: %.6f", m.Coordenadas.Latitud)
	}

	// Validar coordenadas.longitud: -180 a 180
	if m.Coordenadas.Longitud < -180 || m.Coordenadas.Longitud > 180 {
		return fmt.Errorf("coordenadas.longitud debe estar entre -180 y 180, recibido: %.6f", m.Coordenadas.Longitud)
	}

	// Validar estado: en_ruta, trabajando, en_pausa, finalizado
	switch m.Estado {
	case "en_ruta", "trabajando", "en_pausa", "finalizado":
		// Estado válido
	default:
		return fmt.Errorf("estado debe ser uno de: en_ruta, trabajando, en_pausa, finalizado, recibido: %s", m.Estado)
	}

	// Validar procentajeProgreso: 0-100
	if m.PorcentajeProgreso < 0 || m.PorcentajeProgreso > 100 {
		return fmt.Errorf("procentajeProgreso debe estar entre 0 y 100, recibido: %d", m.PorcentajeProgreso)
	}

	// Validar nivelBateria: 0-100
	if m.NivelBateria < 0 || m.NivelBateria > 100 {
		return fmt.Errorf("nivelBateria debe estar entre 0 y 100, recibido: %d", m.NivelBateria)
	}

	return nil
}

// EventoInventarioCuadrilla representa el evento publicado a NATS.
type EventoInventarioCuadrilla struct {
	GrupoTrabajo       string      `json:"grupo_trabajo"`
	NombreEmpleado     string      `json:"nombre_empleado"`
	Timestamp          time.Time   `json:"timestamp"`
	Coordenadas        Coordenadas `json:"coordenadas"`
	CodigoODT          string      `json:"codigo_odt"`
	Estado             string      `json:"estado"`
	PorcentajeProgreso int         `json:"porcentaje_progreso"`
	NivelBateria       int         `json:"nivel_bateria"`
	RecibidoEn         time.Time   `json:"recibido_en"`
}
