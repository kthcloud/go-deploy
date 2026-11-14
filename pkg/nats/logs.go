package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	defaultLogsStream = "logs"
	defaultLogsMaxAge = 24 * time.Hour
)

// LogMessage matches the JSON payload published by Vector and go-deploy into NATS.
type LogMessage struct {
	Source    string    `json:"source"`
	Prefix    string    `json:"prefix"`
	Line      string    `json:"line"`
	CreatedAt time.Time `json:"createdAt"`
}

// LogsStreamName returns the configured JetStream stream name for deployment logs.
func LogsStreamName() string {
	if s := config.Config.NATS.Logs.Stream; s != "" {
		return s
	}
	return defaultLogsStream
}

// LogSubject returns the NATS subject for a deployment's logs in a given zone.
func LogSubject(zoneName, namespace, deployName string) string {
	return fmt.Sprintf("logs.%s.%s.%s", zoneName, namespace, deployName)
}

// EnsureLogsStream creates or updates the JetStream stream used for deployment logs.
func (c *Client) EnsureLogsStream(ctx context.Context) error {
	js := c.JetStream()
	if js == nil {
		return fmt.Errorf("nats jetstream is not initialized")
	}

	maxAge := defaultLogsMaxAge
	if raw := config.Config.NATS.Logs.MaxAge; raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return fmt.Errorf("invalid nats.logs.maxAge %q: %w", raw, err)
		}
		maxAge = parsed
	}

	streamName := LogsStreamName()
	_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{"logs.>"},
		Storage:  jetstream.FileStorage,
		MaxAge:   maxAge,
	})
	if err != nil {
		return fmt.Errorf("create or update stream %q: %w", streamName, err)
	}

	return nil
}

// PublishLog publishes a log entry to the deployment's NATS JetStream subject.
func (c *Client) PublishLog(ctx context.Context, zoneName, namespace, deployName string, msg LogMessage) error {
	js := c.JetStream()
	if js == nil {
		return fmt.Errorf("nats jetstream is not initialized")
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal log message: %w", err)
	}

	if _, err := js.Publish(ctx, LogSubject(zoneName, namespace, deployName), payload); err != nil {
		return fmt.Errorf("publish log: %w", err)
	}
	return nil
}

// StreamLogs replays up to `history` recent log lines from JetStream and then streams live logs
// until ctx is canceled.
func (c *Client) StreamLogs(
	ctx context.Context,
	zoneName, namespace, deployName string,
	history int,
	handler func(source, prefix, line string, createdAt time.Time),
) error {
	js := c.JetStream()
	if js == nil {
		return fmt.Errorf("nats jetstream is not initialized")
	}

	deliverPolicy := jetstream.DeliverAllPolicy
	if history <= 0 {
		deliverPolicy = jetstream.DeliverNewPolicy
	}

	cons, err := js.OrderedConsumer(ctx, LogsStreamName(), jetstream.OrderedConsumerConfig{
		FilterSubjects: []string{LogSubject(zoneName, namespace, deployName)},
		DeliverPolicy:  deliverPolicy,
	})
	if err != nil {
		return fmt.Errorf("create ordered consumer: %w", err)
	}

	msgs, err := cons.Messages()
	if err != nil {
		return fmt.Errorf("start message iterator: %w", err)
	}
	defer msgs.Stop()

	go func() {
		<-ctx.Done()
		msgs.Stop()
	}()

	replaying := history > 0
	var backlog []LogMessage

	for {
		msg, err := msgs.Next()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, jetstream.ErrMsgIteratorClosed) {
				return nil
			}
			return fmt.Errorf("read next log message: %w", err)
		}

		var entry LogMessage
		if err := json.Unmarshal(msg.Data(), &entry); err != nil {
			continue
		}

		if replaying {
			backlog = append(backlog, entry)
			if len(backlog) > history {
				backlog = backlog[len(backlog)-history:]
			}

			meta, metaErr := msg.Metadata()
			if metaErr == nil && meta.NumPending > 0 {
				continue
			}

			replaying = false
			for _, item := range backlog {
				handler(item.Source, item.Prefix, item.Line, item.CreatedAt)
			}
			backlog = nil
			continue
		}

		handler(entry.Source, entry.Prefix, entry.Line, entry.CreatedAt)
	}
}
