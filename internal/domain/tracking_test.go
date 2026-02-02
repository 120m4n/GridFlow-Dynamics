package domain

import (
	"testing"
	"time"
)

func TestEstadoCuadrilla(t *testing.T) {
	estados := []EstadoCuadrilla{
		EstadoEnRuta,
		EstadoTrabajando,
		EstadoEnPausa,
		EstadoFinalizado,
	}

	for _, estado := range estados {
		if estado == "" {
			t.Error("Estado no debe estar vacío")
		}
	}
}

func TestCoordenadasStruct(t *testing.T) {
	coords := Coordenadas{
		Latitud:  40.7128,
		Longitud: -74.0060,
	}

	if coords.Latitud != 40.7128 {
		t.Errorf("Latitud = %f; esperado 40.7128", coords.Latitud)
	}

	if coords.Longitud != -74.0060 {
		t.Errorf("Longitud = %f; esperado -74.0060", coords.Longitud)
	}
}

func TestMensajeInventarioCuadrillaValidar(t *testing.T) {
	tests := []struct {
		nombre      string
		mensaje     MensajeInventarioCuadrilla
		debeErrorar bool
		errorMsg    string
	}{
		{
			nombre: "Mensaje válido",
			mensaje: MensajeInventarioCuadrilla{
				GrupoTrabajo:       "G0/CUADRILLA_123",
				NombreEmpleado:     "Juan Perez",
				Timestamp:          time.Now(),
				Coordenadas:        Coordenadas{Latitud: 40.7128, Longitud: -74.0060},
				CodigoODT:          "ODT-001",
				Estado:             "trabajando",
				PorcentajeProgreso: 75,
				NivelBateria:       85,
			},
			debeErrorar: false,
		},
		{
			nombre: "GrupoTrabajo vacío",
			mensaje: MensajeInventarioCuadrilla{
				GrupoTrabajo:       "",
				NombreEmpleado:     "Juan Perez",
				Timestamp:          time.Now(),
				Coordenadas:        Coordenadas{Latitud: 40.7128, Longitud: -74.0060},
				CodigoODT:          "ODT-001",
				Estado:             "trabajando",
				PorcentajeProgreso: 75,
				NivelBateria:       85,
			},
			debeErrorar: true,
			errorMsg:    "grupoTrabajo es requerido",
		},
		{
			nombre: "NombreEmpleado vacío",
			mensaje: MensajeInventarioCuadrilla{
				GrupoTrabajo:       "G0/CUADRILLA_123",
				NombreEmpleado:     "",
				Timestamp:          time.Now(),
				Coordenadas:        Coordenadas{Latitud: 40.7128, Longitud: -74.0060},
				CodigoODT:          "ODT-001",
				Estado:             "trabajando",
				PorcentajeProgreso: 75,
				NivelBateria:       85,
			},
			debeErrorar: true,
			errorMsg:    "nombreEmpleado es requerido",
		},
		{
			nombre: "Latitud inválida",
			mensaje: MensajeInventarioCuadrilla{
				GrupoTrabajo:       "G0/CUADRILLA_123",
				NombreEmpleado:     "Juan Perez",
				Timestamp:          time.Now(),
				Coordenadas:        Coordenadas{Latitud: 100.0, Longitud: -74.0060},
				CodigoODT:          "ODT-001",
				Estado:             "trabajando",
				PorcentajeProgreso: 75,
				NivelBateria:       85,
			},
			debeErrorar: true,
			errorMsg:    "latitud debe estar entre -90 y 90",
		},
		{
			nombre: "Estado inválido",
			mensaje: MensajeInventarioCuadrilla{
				GrupoTrabajo:       "G0/CUADRILLA_123",
				NombreEmpleado:     "Juan Perez",
				Timestamp:          time.Now(),
				Coordenadas:        Coordenadas{Latitud: 40.7128, Longitud: -74.0060},
				CodigoODT:          "ODT-001",
				Estado:             "estado_invalido",
				PorcentajeProgreso: 75,
				NivelBateria:       85,
			},
			debeErrorar: true,
			errorMsg:    "estado debe ser uno de",
		},
		{
			nombre: "Porcentaje de progreso inválido",
			mensaje: MensajeInventarioCuadrilla{
				GrupoTrabajo:       "G0/CUADRILLA_123",
				NombreEmpleado:     "Juan Perez",
				Timestamp:          time.Now(),
				Coordenadas:        Coordenadas{Latitud: 40.7128, Longitud: -74.0060},
				CodigoODT:          "ODT-001",
				Estado:             "trabajando",
				PorcentajeProgreso: 150,
				NivelBateria:       85,
			},
			debeErrorar: true,
			errorMsg:    "procentajeProgreso debe estar entre 0 y 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.nombre, func(t *testing.T) {
			err := tt.mensaje.Validar()
			if tt.debeErrorar {
				if err == nil {
					t.Error("Se esperaba un error pero no se obtuvo ninguno")
				}
			} else {
				if err != nil {
					t.Errorf("Error inesperado: %v", err)
				}
			}
		})
	}
}

func TestEventoInventarioCuadrilla(t *testing.T) {
	evento := EventoInventarioCuadrilla{
		GrupoTrabajo:       "G0/CUADRILLA_123",
		NombreEmpleado:     "Juan Perez",
		Timestamp:          time.Now(),
		Coordenadas:        Coordenadas{Latitud: 40.7128, Longitud: -74.0060},
		CodigoODT:          "ODT-001",
		Estado:             "trabajando",
		PorcentajeProgreso: 75,
		NivelBateria:       85,
		RecibidoEn:         time.Now(),
	}

	if evento.GrupoTrabajo == "" {
		t.Error("GrupoTrabajo no debe estar vacío")
	}

	if evento.Estado != "trabajando" {
		t.Errorf("Estado = %s; esperado trabajando", evento.Estado)
	}
}
