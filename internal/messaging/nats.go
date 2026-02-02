package messaging
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

















































































}	return nil	// No hay recursos específicos del publisher para cerrarfunc (p *Publisher) Close() error {// Close cierra el publisher.}	return nil	log.Printf("Evento publicado a subject '%s'", subject)	}		return fmt.Errorf("fallo al publicar mensaje: %w", err)	if err := p.conn.conn.Publish(subject, payload); err != nil {	}		return fmt.Errorf("fallo al serializar mensaje: %w", err)	if err != nil {	payload, err := json.Marshal(data)func (p *Publisher) Publish(ctx context.Context, subject string, data interface{}) error {// Publish publica un mensaje a un subject específico.}	return &Publisher{conn: conn}, nil	}		return nil, fmt.Errorf("conexión NATS no está activa")	if !conn.IsConnected() {func NewPublisher(conn *Connection) (*Publisher, error) {// NewPublisher crea un nuevo publisher.}	conn *Connectiontype Publisher struct {// Publisher publica eventos a NATS.}	return c.connfunc (c *Connection) GetConn() *nats.Conn {// GetConn retorna la conexión nativa de NATS.}	return c.conn != nil && c.conn.IsConnected()func (c *Connection) IsConnected() bool {// IsConnected retorna si la conexión está activa.}	return nil	}		log.Println("Conexión NATS cerrada")		c.conn.Close()	if c.conn != nil {func (c *Connection) Close() error {// Close cierra la conexión NATS.}	return nil	log.Printf("Conectado a NATS en %s", c.url)	c.conn = conn	}		return fmt.Errorf("fallo al conectar a NATS: %w", err)	if err != nil {	conn, err := nats.Connect(c.url, opts...)	}		}),			log.Printf("NATS reconectado a %s", nc.ConnectedUrl())		nats.ReconnectHandler(func(nc *nats.Conn) {		}),			}				log.Printf("NATS desconectado: %v", err)			if err != nil {		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {		nats.ReconnectWait(2 * time.Second),		nats.MaxReconnects(-1),		nats.Name("GridFlow-Dynamics"),	opts := []nats.Option{func (c *Connection) Connect() error {// Connect establece la conexión con NATS.}	}