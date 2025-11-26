package messaging

import (
	"testing"
)

func TestExchangeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"ExchangeCrewEvents", ExchangeCrewEvents, "crew.events"},
		{"ExchangeTaskEvents", ExchangeTaskEvents, "task.events"},
		{"ExchangeAlertEvents", ExchangeAlertEvents, "alert.events"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %s; want %s", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestRoutingKeyConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"RoutingKeyCrewLocation", RoutingKeyCrewLocation, "crew.location.update"},
		{"RoutingKeyCrewStatus", RoutingKeyCrewStatus, "crew.status.update"},
		{"RoutingKeyTaskStatus", RoutingKeyTaskStatus, "task.status.update"},
		{"RoutingKeyAlertCreated", RoutingKeyAlertCreated, "alert.created"},
		{"RoutingKeyAlertAcked", RoutingKeyAlertAcked, "alert.acknowledged"},
		{"RoutingKeyAlertResolved", RoutingKeyAlertResolved, "alert.resolved"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %s; want %s", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestNewConnection(t *testing.T) {
	url := "amqp://guest:guest@localhost:5672/"
	conn := NewConnection(url)

	if conn == nil {
		t.Fatal("NewConnection returned nil")
	}

	if conn.url != url {
		t.Errorf("Connection.url = %s; want %s", conn.url, url)
	}

	if conn.connected {
		t.Error("Connection should not be connected initially")
	}

	if conn.closeChan == nil {
		t.Error("Connection.closeChan should not be nil")
	}
}

func TestConnectionIsConnected(t *testing.T) {
	conn := NewConnection("amqp://guest:guest@localhost:5672/")

	if conn.IsConnected() {
		t.Error("IsConnected should return false before Connect is called")
	}
}

func TestConnectionConnectFails(t *testing.T) {
	// Use an invalid URL to test connection failure
	conn := NewConnection("amqp://invalid:invalid@nonexistent:5672/")
	err := conn.Connect()

	if err == nil {
		t.Error("Connect should fail with invalid URL")
		conn.Close()
	}
}
