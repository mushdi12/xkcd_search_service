package nats

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go"
	"yadro.com/course/search/core"
)

type natsConn interface {
	Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error)
	Drain() error
}

type Listener struct {
	nc        natsConn
	log       *slog.Logger
	initiator core.Initiator
	topic     string
}

func NewListener(address string, topic string, log *slog.Logger, initiator core.Initiator) (*Listener, error) {
	nc, err := nats.Connect(address)
	if err != nil {
		return nil, err
	}
	return &Listener{nc: nc, log: log, initiator: initiator, topic: topic}, nil
}

func (l *Listener) Listen(ctx context.Context) {
	_, err := l.nc.Subscribe(l.topic, func(msg *nats.Msg) {
		message := string(msg.Data)
		l.log.Info("received message", "topic", l.topic, "data", message)

		switch message {
		case "update":
			l.log.Info("handling update event, rebuilding index")
			if err := l.initiator.IndexComics(ctx); err != nil {
				l.log.Info("failed to rebuild index", "error", err)
			}
		case "drop":
			l.log.Info("handling drop event, clearing index")
			if err := l.initiator.ClearIndex(ctx); err != nil {
				l.log.Info("failed to clear index", "error", err)
			}
		}

	})

	if err != nil {
		l.log.Info("failed to subscribe", "error", err)
	}
}

func (l *Listener) Close() error {
	return l.nc.Drain()
}
