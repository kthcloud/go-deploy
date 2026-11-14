package nats

import (
	"context"
	"testing"
	"time"

	"github.com/kthcloud/go-deploy/models/config"
	configPkg "github.com/kthcloud/go-deploy/pkg/config"
)

func TestLogSubjectAndStreamName(t *testing.T) {
	if got := LogSubject("local", "deploy", "my-app"); got != "logs.local.deploy.my-app" {
		t.Fatalf("unexpected subject: %q", got)
	}

	orig := configPkg.Config.NATS.Logs.Stream
	defer func() { configPkg.Config.NATS.Logs.Stream = orig }()

	configPkg.Config.NATS.Logs.Stream = ""
	if got := LogsStreamName(); got != "logs" {
		t.Fatalf("expected default stream 'logs', got %q", got)
	}

	configPkg.Config.NATS.Logs.Stream = "custom-logs"
	if got := LogsStreamName(); got != "custom-logs" {
		t.Fatalf("expected 'custom-logs', got %q", got)
	}
}

func TestIsZoneLogsEnabled(t *testing.T) {
	origNATS := configPkg.Config.NATS
	origZones := configPkg.Config.Zones
	defer func() {
		configPkg.Config.NATS = origNATS
		configPkg.Config.Zones = origZones
	}()

	configPkg.Config.NATS.URL = "nats://localhost:4222"
	configPkg.Config.NATS.Logs.Enabled = true
	configPkg.Config.Zones = []config.Zone{
		{Name: "new-zone", Enabled: true, Capabilities: []string{config.ZoneCapabilityDeployment, config.ZoneCapabilityNatsLogs}},
		{Name: "old-zone", Enabled: true, Capabilities: []string{config.ZoneCapabilityDeployment}},
	}

	if !IsZoneLogsEnabled("new-zone") {
		t.Fatal("expected nats logs enabled for zone with nats-logs capability")
	}
	if IsZoneLogsEnabled("old-zone") {
		t.Fatal("expected fallback to legacy logs for zone without nats-logs capability")
	}
	if IsZoneLogsEnabled("missing-zone") {
		t.Fatal("expected false for unknown zone")
	}
}

func TestUninitializedClientErrors(t *testing.T) {
	c := &Client{}
	ctx := context.Background()

	if err := c.EnsureLogsStream(ctx); err == nil {
		t.Fatal("expected error from EnsureLogsStream on uninitialized client")
	}
	if err := c.PublishLog(ctx, "local", "deploy", "app", LogMessage{Line: "hi", CreatedAt: time.Now()}); err == nil {
		t.Fatal("expected error from PublishLog on uninitialized client")
	}
	if err := c.StreamLogs(ctx, "local", "deploy", "app", 10, func(string, string, string, time.Time) {}); err == nil {
		t.Fatal("expected error from StreamLogs on uninitialized client")
	}
}
