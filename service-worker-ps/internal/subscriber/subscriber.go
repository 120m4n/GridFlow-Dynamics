// Package subscriber provides NATS message subscription functionality.
package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Message represents the structure of messages received from NATS.
// This matches the EventoInventarioCuadrilla structure from the main API.
type Message struct {
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

// Coordenadas represents GPS coordinates.
type Coordenadas struct {
	Latitud  float64 `json:"latitud"`
	Longitud float64 `json:"longitud"`
}

// MessageHandler is a function that processes received messages.
type MessageHandler func(ctx context.Context, msg *Message) error

// Subscriber manages NATS subscriptions and message handling.
type Subscriber struct {
	conn    *nats.Conn
	sub     *nats.Subscription
	handler MessageHandler
}

// NewSubscriber creates a new NATS subscriber.
func NewSubscriber(natsURL string, handler MessageHandler) (*Subscriber, error) {
	opts := []nats.Option{
		nats.Name("ServiceWorker-PS"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Printf("NATS disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
	}

	conn, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	log.Printf("Connected to NATS at %s", natsURL)

	return &Subscriber{
		conn:    conn,
		handler: handler,
	}, nil
}

// Subscribe starts listening to the specified NATS subject.
func (s *Subscriber) Subscribe(subject string) error {
	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.processMessage(msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s: %w", subject, err)
	}

	s.sub = sub
	log.Printf("Subscribed to NATS subject: %s", subject)
	return nil
}

// processMessage handles incoming NATS messages.
func (s *Subscriber) processMessage(msg *nats.Msg) {
	var message Message
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.handler(ctx, &message); err != nil {
		log.Printf("Failed to handle message: %v", err)
		return
	}

	log.Printf("Successfully processed message from cuadrilla: %s", message.GrupoTrabajo)
}

// Close closes the NATS connection and unsubscribes.
func (s *Subscriber) Close() error {
	if s.sub != nil {
		if err := s.sub.Unsubscribe(); err != nil {
			log.Printf("Warning: failed to unsubscribe: %v", err)
		}
	}

	if s.conn != nil {
		s.conn.Close()
		log.Println("NATS connection closed")
	}

	return nil
}

// IsConnected returns whether the NATS connection is active.
func (s *Subscriber) IsConnected() bool {
	return s.conn != nil && s.conn.IsConnected()
}
