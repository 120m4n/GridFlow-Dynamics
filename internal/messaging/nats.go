// Package messaging provides NATS messaging infrastructure for event-driven communication.
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Subjects para la arquitectura orientada a eventos.
const (
	SubjectInventarioCuadrilla = "inventario.cuadrilla"
)

// Connection representa una conexión a NATS con soporte de reconexión.
type Connection struct {
	url  string
	conn *nats.Conn
}

// NewConnection crea una nueva conexión NATS.
func NewConnection(url string) *Connection {
	return &Connection{
		url: url,
	}
}

// Connect establece la conexión con NATS.
func (c *Connection) Connect() error {
	opts := []nats.Option{
		nats.Name("GridFlow-Dynamics"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Printf("NATS desconectado: %v", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconectado a %s", nc.ConnectedUrl())
		}),
	}

	conn, err := nats.Connect(c.url, opts...)
	if err != nil {
		return fmt.Errorf("fallo al conectar a NATS: %w", err)
	}

	c.conn = conn
	log.Printf("Conectado a NATS en %s", c.url)
	return nil
}

// Close cierra la conexión NATS.
func (c *Connection) Close() error {
	if c.conn != nil {
		c.conn.Close()
		log.Println("Conexión NATS cerrada")
	}
	return nil
}

// IsConnected retorna si la conexión está activa.
func (c *Connection) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

// GetConn retorna la conexión nativa de NATS.
func (c *Connection) GetConn() *nats.Conn {
	return c.conn
}

// Publisher publica eventos a NATS.
type Publisher struct {
	conn *Connection
}

// NewPublisher crea un nuevo publisher.
func NewPublisher(conn *Connection) (*Publisher, error) {
	if !conn.IsConnected() {
		return nil, fmt.Errorf("conexión NATS no está activa")
	}
	return &Publisher{conn: conn}, nil
}

// Publish publica un mensaje a un subject específico.
func (p *Publisher) Publish(ctx context.Context, subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("fallo al serializar mensaje: %w", err)
	}

	if err := p.conn.conn.Publish(subject, payload); err != nil {
		return fmt.Errorf("fallo al publicar mensaje: %w", err)
	}

	log.Printf("Evento publicado a subject '%s'", subject)
	return nil
}

// Close cierra el publisher.
func (p *Publisher) Close() error {
	return nil
}
