package grpc

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
)

type fakeUpdater struct {
	status    core.ServiceStatus
	stats     core.ServiceStats
	statsErr  error
	updateErr error
	dropErr   error
}

func (f fakeUpdater) Update(ctx context.Context) error {
	return f.updateErr
}

func (f fakeUpdater) Stats(ctx context.Context) (core.ServiceStats, error) {
	return f.stats, f.statsErr
}

func (f fakeUpdater) Status(ctx context.Context) core.ServiceStatus {
	return f.status
}

func (f fakeUpdater) Drop(ctx context.Context) error {
	return f.dropErr
}

func TestServer_Ping(t *testing.T) {
	s := NewServer(fakeUpdater{})
	_, err := s.Ping(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("Ping returned error: %v", err)
	}
}

func TestServer_Status_Running(t *testing.T) {
	s := NewServer(fakeUpdater{status: core.StatusRunning})
	reply, err := s.Status(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if reply.Status != updatepb.Status_STATUS_RUNNING {
		t.Fatalf("expected STATUS_RUNNING, got %v", reply.Status)
	}
}

func TestServer_Status_Idle(t *testing.T) {
	s := NewServer(fakeUpdater{status: core.StatusIdle})
	reply, err := s.Status(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if reply.Status != updatepb.Status_STATUS_IDLE {
		t.Fatalf("expected STATUS_IDLE, got %v", reply.Status)
	}
}

func TestServer_Update_Success(t *testing.T) {
	s := NewServer(fakeUpdater{})
	_, err := s.Update(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

func TestServer_Update_AlreadyExists(t *testing.T) {
	s := NewServer(fakeUpdater{updateErr: core.ErrAlreadyExists})
	_, err := s.Update(context.Background(), &emptypb.Empty{})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}
}

func TestServer_Stats_Success(t *testing.T) {
	s := NewServer(fakeUpdater{
		stats: core.ServiceStats{
			DBStats: core.DBStats{
				WordsTotal:    100,
				WordsUnique:   50,
				ComicsFetched: 150,
			},
			ComicsTotal: 200,
		},
	})
	reply, err := s.Stats(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("Stats returned error: %v", err)
	}
	if reply.WordsTotal != 100 || reply.ComicsTotal != 200 {
		t.Fatalf("unexpected stats: %#v", reply)
	}
}

func TestServer_Stats_Error(t *testing.T) {
	s := NewServer(fakeUpdater{statsErr: errors.New("db error")})
	_, err := s.Stats(context.Background(), &emptypb.Empty{})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", err)
	}
}

func TestServer_Drop_Success(t *testing.T) {
	s := NewServer(fakeUpdater{})
	_, err := s.Drop(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("Drop returned error: %v", err)
	}
}

func TestServer_Drop_Error(t *testing.T) {
	s := NewServer(fakeUpdater{dropErr: errors.New("db error")})
	_, err := s.Drop(context.Background(), &emptypb.Empty{})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", err)
	}
}
