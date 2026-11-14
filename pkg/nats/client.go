package nats

import (
	"context"
	"fmt"
	"sync"

	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Client wraps a NATS connection and JetStream context.
type Client struct {
	mu   sync.RWMutex
	conn *nats.Conn
	js   jetstream.JetStream
}

// Default is the package-level NATS client initialized by Setup.
var Default = &Client{}

// IsEnabled reports whether NATS is configured.
func IsEnabled() bool {
	return config.Config.NATS.URL != ""
}

// IsLogsEnabled reports whether NATS JetStream deployment logging is enabled.
func IsLogsEnabled() bool {
	return IsEnabled() && config.Config.NATS.Logs.Enabled
}

// IsZoneLogsEnabled reports whether NATS JetStream deployment logging is enabled for the given zone.
func IsZoneLogsEnabled(zoneName string) bool {
	if !IsLogsEnabled() {
		return false
	}
	z := config.Config.GetZone(zoneName)
	return z != nil && z.HasCapability("nats-logs")
}

// Setup connects the default NATS client and provisions required JetStream streams.
func Setup(ctx context.Context) error {
	if !IsEnabled() {
		return nil
	}

	cfg := config.Config.NATS
	var opts []nats.Option
	if cfg.Username != "" && cfg.Password != "" {
		opts = append(opts, nats.UserInfo(cfg.Username, cfg.Password))
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return fmt.Errorf("connect nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return fmt.Errorf("init jetstream: %w", err)
	}

	Default.mu.Lock()
	Default.conn = nc
	Default.js = js
	Default.mu.Unlock()

	if cfg.Logs.Enabled {
		if err := Default.EnsureLogsStream(ctx); err != nil {
			_ = Shutdown(ctx)
			return fmt.Errorf("ensure logs stream: %w", err)
		}
	}

	log.Println("Connected to NATS", cfg.URL)
	return nil
}

// Shutdown drains and closes the default NATS connection.
func Shutdown(_ context.Context) error {
	Default.mu.Lock()
	nc := Default.conn
	Default.conn = nil
	Default.js = nil
	Default.mu.Unlock()

	if nc == nil {
		return nil
	}

	if err := nc.Drain(); err != nil {
		nc.Close()
		return fmt.Errorf("drain nats: %w", err)
	}
	return nil
}

// JetStream returns the configured JetStream client, if connected.
func (c *Client) JetStream() jetstream.JetStream {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.js
}

// Conn returns the underlying NATS connection, if connected.
func (c *Client) Conn() *nats.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}
