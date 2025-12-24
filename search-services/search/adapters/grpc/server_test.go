package grpc

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

type fakeSearcher struct {
	searchResult      []core.Comics
	searchErr         error
	indexSearchResult []core.Comics
	indexSearchErr    error
}

func (f fakeSearcher) Search(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	return f.searchResult, f.searchErr
}

func (f fakeSearcher) IndexSearch(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	return f.indexSearchResult, f.indexSearchErr
}

func TestServer_Ping(t *testing.T) {
	s := NewServer(fakeSearcher{})

	resp, err := s.Ping(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("Ping returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
}

func TestServer_Search_DefaultLimitAndMapping(t *testing.T) {
	s := NewServer(fakeSearcher{
		searchResult: []core.Comics{
			{ID: 1, URL: "url1"},
			{ID: 2, URL: "url2"},
		},
	})

	req := &searchpb.SearchRequest{
		Phrase: "linux",
		Limit:  0, 
	}

	resp, err := s.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(resp.Comics) != 2 {
		t.Fatalf("expected 2 comics, got %d", len(resp.Comics))
	}
	if resp.Comics[0].Id != 1 || resp.Comics[0].Url != "url1" {
		t.Fatalf("unexpected first comics: %#v", resp.Comics[0])
	}
}

func TestServer_Search_NotFound(t *testing.T) {
	s := NewServer(fakeSearcher{
		searchErr: core.ErrNotFound,
	})

	_, err := s.Search(context.Background(), &searchpb.SearchRequest{Phrase: "unknown"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T", err)
	}
	if st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", st.Code())
	}
}

func TestServer_Search_OtherError(t *testing.T) {
	s := NewServer(fakeSearcher{
		searchErr: errors.New("some error"),
	})

	_, err := s.Search(context.Background(), &searchpb.SearchRequest{Phrase: "linux"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if _, ok := status.FromError(err); ok {
		t.Fatalf("expected raw error, got grpc status")
	}
}

func TestServer_IndexSearch_DefaultLimitAndMapping(t *testing.T) {
	s := NewServer(fakeSearcher{
		indexSearchResult: []core.Comics{
			{ID: 10, URL: "url10"},
		},
	})

	req := &searchpb.SearchRequest{
		Phrase: "linux",
		Limit:  0,
	}

	resp, err := s.IndexSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("IndexSearch returned error: %v", err)
	}
	if len(resp.Comics) != 1 {
		t.Fatalf("expected 1 comics, got %d", len(resp.Comics))
	}
	if resp.Comics[0].Id != 10 || resp.Comics[0].Url != "url10" {
		t.Fatalf("unexpected comics: %#v", resp.Comics[0])
	}
}

func TestServer_IndexSearch_NotFound(t *testing.T) {
	s := NewServer(fakeSearcher{
		indexSearchErr: core.ErrNotFound,
	})

	_, err := s.IndexSearch(context.Background(), &searchpb.SearchRequest{Phrase: "unknown"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T", err)
	}
	if st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", st.Code())
	}
}

func TestServer_IndexSearch_OtherError(t *testing.T) {
	s := NewServer(fakeSearcher{
		indexSearchErr: errors.New("some error"),
	})

	_, err := s.IndexSearch(context.Background(), &searchpb.SearchRequest{Phrase: "linux"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if _, ok := status.FromError(err); ok {
		t.Fatalf("expected raw error, got grpc status")
	}
}


