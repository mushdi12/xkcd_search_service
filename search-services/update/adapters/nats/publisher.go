package nats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"yadro.com/course/update/core"
)

const topic = "xkcd.db.updated"

type natsConn interface {
	Publish(subj string, data []byte) error
	Drain() error
}

type Notificator struct {
	nc  natsConn
	log *slog.Logger
}

func NewNotificator(address string, log *slog.Logger) (*Notificator, error) {
	nc, err := nats.Connect(address)
	if err != nil {
		return nil, err
	}
	return &Notificator{nc: nc, log: log}, nil
}

func (n *Notificator) Publish(ctx context.Context, event core.EventType) error {
	err := n.nc.Publish(topic, []byte(event))
	if err != nil {
		n.log.Error("failed to publish message", "topic", topic, "error", err)
		return fmt.Errorf("failed to publish message: %v", err)
	}

	n.log.Info("message published", "topic", topic)
	return nil
}

func (n *Notificator) Close() error {
	return n.nc.Drain()
}
