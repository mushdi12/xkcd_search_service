package update

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	updatepb "yadro.com/course/proto/update"
)

type fakeUpdateClient struct {
	pingErr   error
	statusRep *updatepb.StatusReply
	statusErr error
	statsRep  *updatepb.StatsReply
	statsErr  error
	updateErr error
	dropErr   error
}

func (f fakeUpdateClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, f.pingErr
}

func (f fakeUpdateClient) Status(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error) {
	return f.statusRep, f.statusErr
}

func (f fakeUpdateClient) Stats(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error) {
	return f.statsRep, f.statsErr
}

func (f fakeUpdateClient) Update(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, f.updateErr
}

func (f fakeUpdateClient) Drop(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, f.dropErr
}

func newUpdateTestClient(f fakeUpdateClient) *Client {
	logger := slog.Default()
	return &Client{
		log:    logger,
		client: f,
	}
}

var _ updatepb.UpdateClient = fakeUpdateClient{}

func TestClient_Status_Mapping(t *testing.T) {
	c := newUpdateTestClient(fakeUpdateClient{
		statusRep: &updatepb.StatusReply{Status: updatepb.Status_STATUS_RUNNING},
	})

	st, err := c.Status(context.Background())
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if st != core.StatusUpdateRunning {
		t.Fatalf("expected running, got %v", st)
	}
}

func TestClient_Status_Error(t *testing.T) {
	c := newUpdateTestClient(fakeUpdateClient{
		statusErr: errors.New("err"),
	})

	st, err := c.Status(context.Background())
	if err == nil || st != core.StatusUpdateUnknown {
		t.Fatalf("expected unknown with error, got st=%v err=%v", st, err)
	}
}

func TestClient_Stats_Mapping(t *testing.T) {
	c := newUpdateTestClient(fakeUpdateClient{
		statsRep: &updatepb.StatsReply{
			WordsTotal:    1,
			WordsUnique:   2,
			ComicsFetched: 3,
			ComicsTotal:   4,
		},
	})

	s, err := c.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats returned error: %v", err)
	}
	if s.ComicsTotal != 4 || s.WordsUnique != 2 {
		t.Fatalf("unexpected stats: %#v", s)
	}
}

func TestClient_Update_AlreadyExists(t *testing.T) {
	c := newUpdateTestClient(fakeUpdateClient{
		updateErr: status.Error(codes.AlreadyExists, "already"),
	})

	err := c.Update(context.Background())
	if !errors.Is(err, core.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestClient_Update_OtherError(t *testing.T) {
	c := newUpdateTestClient(fakeUpdateClient{
		updateErr: errors.New("err"),
	})

	err := c.Update(context.Background())
	if err == nil || errors.Is(err, core.ErrAlreadyExists) {
		t.Fatalf("expected passthrough error, got %v", err)
	}
}

func TestClient_PingAndDrop(t *testing.T) {
	c := newUpdateTestClient(fakeUpdateClient{})

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping returned error: %v", err)
	}
	if err := c.Drop(context.Background()); err != nil {
		t.Fatalf("Drop returned error: %v", err)
	}
}

