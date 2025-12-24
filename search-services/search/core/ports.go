package core

import (
	"context"
)

type Storager interface {
	Search(ctx context.Context, keyword string) ([]int, error)
	Get(ctx context.Context, ID int) (Comics, error)
	GetAllComics(ctx context.Context) ([]Comics, error)
	GetComicsByIDs(ctx context.Context, ids ...int) ([]Comics, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type Searcher interface {
	Search(context.Context, string, int) ([]Comics, error)
	IndexSearch(context.Context, string, int) ([]Comics, error)
}

type Initiator interface {
	GetIndexedComics(ctx context.Context, words []string, limit int) ([]Comics, error)
	IndexComics(ctx context.Context) error
	ClearIndex(ctx context.Context) error
}

type Notificator interface {
	Subscribe(context.Context, EventType) 
}
