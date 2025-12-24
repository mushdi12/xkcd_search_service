package nats

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"yadro.com/course/update/core"
)

type fakeNATSConn struct {
	publishErr error
	drainErr   error
	published  []struct {
		subject string
		data    []byte
	}
}

func (f *fakeNATSConn) Publish(subj string, data []byte) error {
	f.published = append(f.published, struct {
		subject string
		data    []byte
	}{subj, data})
	return f.publishErr
}

func (f *fakeNATSConn) Drain() error {
	return f.drainErr
}

func TestNotificator_Publish_Success(t *testing.T) {
	fakeConn := &fakeNATSConn{}
	n := &Notificator{
		nc:  fakeConn,
		log: slog.Default(),
	}

	err := n.Publish(context.Background(), core.EventTypeUpdating)
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}
	if len(fakeConn.published) != 1 {
		t.Fatalf("expected 1 publish, got %d", len(fakeConn.published))
	}
	if fakeConn.published[0].subject != topic {
		t.Fatalf("expected subject %q, got %q", topic, fakeConn.published[0].subject)
	}
	if string(fakeConn.published[0].data) != string(core.EventTypeUpdating) {
		t.Fatalf("expected data %q, got %q", core.EventTypeUpdating, string(fakeConn.published[0].data))
	}
}

func TestNotificator_Publish_Error(t *testing.T) {
	fakeConn := &fakeNATSConn{publishErr: errors.New("nats error")}
	n := &Notificator{
		nc:  fakeConn,
		log: slog.Default(),
	}

	err := n.Publish(context.Background(), core.EventTypeDropped)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestNotificator_Close(t *testing.T) {
	fakeConn := &fakeNATSConn{}
	n := &Notificator{
		nc:  fakeConn,
		log: slog.Default(),
	}

	err := n.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}
