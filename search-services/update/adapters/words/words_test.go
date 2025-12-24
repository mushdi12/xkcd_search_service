package words

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
	"yadro.com/course/update/core"
)

type fakeWordsClient struct {
	pingErr error
	normRep *wordspb.WordsReply
	normErr error
}

func (f fakeWordsClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, f.pingErr
}

func (f fakeWordsClient) Norm(ctx context.Context, in *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
	return f.normRep, f.normErr
}

func newWordsTestClient(f fakeWordsClient) *Client {
	logger := slog.Default()
	return &Client{
		log:    logger,
		client: f,
	}
}

var _ wordspb.WordsClient = fakeWordsClient{}

func TestClient_Norm_Success(t *testing.T) {
	c := newWordsTestClient(fakeWordsClient{
		normRep: &wordspb.WordsReply{Words: []string{"a", "b"}},
	})

	res, err := c.Norm(context.Background(), "phrase")
	if err != nil {
		t.Fatalf("Norm returned error: %v", err)
	}
	if len(res) != 2 || res[0] != "a" {
		t.Fatalf("unexpected result: %#v", res)
	}
}

func TestClient_Norm_BadArguments(t *testing.T) {
	c := newWordsTestClient(fakeWordsClient{
		normErr: status.Error(codes.ResourceExhausted, "too long"),
	})

	_, err := c.Norm(context.Background(), "phrase")
	if !errors.Is(err, core.ErrBadArguments) {
		t.Fatalf("expected ErrBadArguments, got %v", err)
	}
}

func TestClient_Ping(t *testing.T) {
	c := newWordsTestClient(fakeWordsClient{})

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping returned error: %v", err)
	}
}
