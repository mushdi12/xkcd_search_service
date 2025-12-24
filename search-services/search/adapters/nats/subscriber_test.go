package nats

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/nats-io/nats.go"
	"yadro.com/course/search/core"
)

type fakeNATSConn struct {
	subscribeErr error
	drainErr     error
	subscribed   []struct {
		subject string
		handler nats.MsgHandler
	}
}

func (f *fakeNATSConn) Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error) {
	f.subscribed = append(f.subscribed, struct {
		subject string
		handler nats.MsgHandler
	}{subj, cb})
	if f.subscribeErr != nil {
		return nil, f.subscribeErr
	}
	return &nats.Subscription{}, nil
}

func (f *fakeNATSConn) Drain() error {
	return f.drainErr
}

type fakeInitiator struct {
	indexErr error
	clearErr error
	indexed  bool
	cleared  bool
}

func (f *fakeInitiator) GetIndexedComics(ctx context.Context, words []string, limit int) ([]core.Comics, error) {
	return nil, nil
}

func (f *fakeInitiator) IndexComics(ctx context.Context) error {
	f.indexed = true
	return f.indexErr
}

func (f *fakeInitiator) ClearIndex(ctx context.Context) error {
	f.cleared = true
	return f.clearErr
}

func TestListener_Listen_Update(t *testing.T) {
	fakeConn := &fakeNATSConn{}
	fakeInit := &fakeInitiator{}
	l := &Listener{
		nc:        fakeConn,
		log:       slog.Default(),
		initiator: fakeInit,
		topic:     "test.topic",
	}

	ctx := context.Background()
	l.Listen(ctx)

	if len(fakeConn.subscribed) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(fakeConn.subscribed))
	}

	msg := &nats.Msg{Data: []byte("update")}
	fakeConn.subscribed[0].handler(msg)

	if !fakeInit.indexed {
		t.Fatalf("expected IndexComics to be called")
	}
}

func TestListener_Listen_Drop(t *testing.T) {
	fakeConn := &fakeNATSConn{}
	fakeInit := &fakeInitiator{}
	l := &Listener{
		nc:        fakeConn,
		log:       slog.Default(),
		initiator: fakeInit,
		topic:     "test.topic",
	}

	ctx := context.Background()
	l.Listen(ctx)

	if len(fakeConn.subscribed) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(fakeConn.subscribed))
	}

	msg := &nats.Msg{Data: []byte("drop")}
	fakeConn.subscribed[0].handler(msg)

	if !fakeInit.cleared {
		t.Fatalf("expected ClearIndex to be called")
	}
}

func TestListener_Listen_SubscribeError(t *testing.T) {
	fakeConn := &fakeNATSConn{subscribeErr: errors.New("nats error")}
	fakeInit := &fakeInitiator{}
	l := &Listener{
		nc:        fakeConn,
		log:       slog.Default(),
		initiator: fakeInit,
		topic:     "test.topic",
	}

	ctx := context.Background()
	l.Listen(ctx)
}

func TestListener_Close(t *testing.T) {
	fakeConn := &fakeNATSConn{}
	l := &Listener{
		nc:  fakeConn,
		log: slog.Default(),
	}

	err := l.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}
