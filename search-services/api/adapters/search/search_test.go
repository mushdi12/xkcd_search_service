package search

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
	searchpb "yadro.com/course/proto/search"
)

type fakeSearchClient struct {
	pingErr        error
	searchReply    *searchpb.SearchReply
	searchErr      error
	indexSearchRep *searchpb.SearchReply
	indexSearchErr error
}

func (f fakeSearchClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, f.pingErr
}

func (f fakeSearchClient) Search(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
	if f.searchReply == nil {
		return nil, f.searchErr
	}
	return f.searchReply, f.searchErr
}

func (f fakeSearchClient) IndexSearch(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
	if f.indexSearchRep == nil {
		return nil, f.indexSearchErr
	}
	return f.indexSearchRep, f.indexSearchErr
}

var _ searchpb.SearchClient = fakeSearchClient{}

func newTestClient(f fakeSearchClient) *Client {
	logger := slog.Default()
	return &Client{
		log:    logger,
		client: f,
	}
}

func TestClient_Search_Success(t *testing.T) {
	reply := &searchpb.SearchReply{
		Comics: []*searchpb.Comics{
			{Id: 1, Url: "u1"},
		},
	}
	c := newTestClient(fakeSearchClient{searchReply: reply})

	res, err := c.Search(context.Background(), "linux", 1)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(res) != 1 || res[0].ID != 1 || res[0].URL != "u1" {
		t.Fatalf("unexpected result: %#v", res)
	}
}

func TestClient_Search_NotFound(t *testing.T) {
	c := newTestClient(fakeSearchClient{
		searchErr: status.Error(codes.NotFound, "not found"),
	})

	_, err := c.Search(context.Background(), "linux", 1)
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestClient_SearchIndex(t *testing.T) {
	reply := &searchpb.SearchReply{
		Comics: []*searchpb.Comics{
			{Id: 2, Url: "u2"},
		},
	}
	c := newTestClient(fakeSearchClient{indexSearchRep: reply})

	res, err := c.SearchIndex(context.Background(), "linux", 1)
	if err != nil {
		t.Fatalf("SearchIndex returned error: %v", err)
	}
	if len(res) != 1 || res[0].ID != 2 {
		t.Fatalf("unexpected result: %#v", res)
	}
}
