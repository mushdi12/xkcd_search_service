package initiator

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"sync"
	"time"

	"yadro.com/course/search/core"
)

type Initiator struct {
	log           *slog.Logger
	indexedComics map[string][]string // id комикса - список слов
	mu            sync.RWMutex
	db            core.Storager
	ttl           time.Duration
	stopCh        chan struct{}
}

func NewInitiator(log *slog.Logger, db core.Storager, ttl time.Duration) *Initiator {

	return &Initiator{
		log:           log,
		db:            db,
		indexedComics: make(map[string][]string),
		ttl:           ttl,
		stopCh:        make(chan struct{}),
	}
}

func (initiator *Initiator) GetIndexedComics(ctx context.Context, words []string, limit int) ([]core.Comics, error) {
	initiator.mu.RLock()
	defer initiator.mu.RUnlock()

	initiator.log.Info("GetIndexedComics called", "words", words, "limit", limit, "indexed_count", len(initiator.indexedComics))

	if len(words) == 0 {
		return []core.Comics{}, nil
	}

	// Создаем set слов для быстрой проверки
	wordsSet := make(map[string]bool)
	for _, word := range words {
		wordsSet[word] = true
	}

	// Подсчитываем релевантность для каждого комикса в индексе
	type comicScore struct {
		id           int
		score        int
		matchedWords int
		totalWords   int
		perfectMatch bool
	}

	scores := make([]comicScore, 0)

	for comicIDStr, comicWords := range initiator.indexedComics {
		comicID, err := strconv.Atoi(comicIDStr)
		if err != nil {
			continue
		}

		matchedCount := 0
		totalMatches := 0
		for _, word := range comicWords {
			if wordsSet[word] {
				matchedCount++
				totalMatches++
			}
		}

		if matchedCount == 0 {
			continue
		}

		perfectMatch := matchedCount == len(words)

		scores = append(scores, comicScore{
			id:           comicID,
			score:        totalMatches,
			matchedWords: matchedCount,
			totalWords:   len(comicWords),
			perfectMatch: perfectMatch,
		})
	}

	if len(scores) == 0 {
		initiator.log.Info("GetIndexedComics no matches", "words", words)
		return []core.Comics{}, nil
	}

	initiator.log.Info("GetIndexedComics found matches", "count", len(scores), "words", words)

	// Ранжирование
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].perfectMatch != scores[j].perfectMatch {
			return scores[i].perfectMatch
		}
		if scores[i].matchedWords != scores[j].matchedWords {
			return scores[i].matchedWords > scores[j].matchedWords
		}
		if scores[i].score != scores[j].score {
			return scores[i].score > scores[j].score
		}
		return scores[i].totalWords < scores[j].totalWords
	})

	comicIDs := make([]int, 0, limit)
	for i, cs := range scores {
		if i >= limit {
			break
		}
		comicIDs = append(comicIDs, cs.id)
	}

	initiator.log.Info("GetIndexedComics fetching comics", "ids", comicIDs)
	comics, err := initiator.db.GetComicsByIDs(ctx, comicIDs...)
	if err != nil {
		initiator.log.Error("GetIndexedComics failed to get comics", "error", err, "ids", comicIDs)
		return nil, fmt.Errorf("failed to get comics by ids: %w", err)
	}
	initiator.log.Info("GetIndexedComics returning", "count", len(comics))
	return comics, nil
}

func (initiator *Initiator) IndexComics(ctx context.Context) error {
	initiator.mu.Lock()
	defer initiator.mu.Unlock()

	comics, err := initiator.db.GetAllComics(ctx)
	if err != nil {
		initiator.log.Error("failed to get all comics", "error", err)
		return err
	}

	// TODO: есть проблема в том, что если комикс удалится из БД, то он останется в
	// индексе (но у нас нет такого функционала вроде)
	for _, comic := range comics {
		comicIDStr := strconv.Itoa(comic.ID)
		initiator.indexedComics[comicIDStr] = comic.Words
	}

	initiator.log.Info("index rebuilt", "comics", len(comics), "indexed", len(initiator.indexedComics))
	return nil
}

func (initiator *Initiator) Start(ctx context.Context) {
	initiator.log.Info("building index immediately on startup")
	if err := initiator.IndexComics(ctx); err != nil {
		initiator.log.Error("failed to build index on startup", "error", err)
	}

	ticker := time.NewTicker(initiator.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			initiator.log.Debug("rebuilding index by timer")
			if err := initiator.IndexComics(ctx); err != nil {
				initiator.log.Error("failed to rebuild index", "error", err)
			}
		case <-initiator.stopCh:
			initiator.log.Info("stopping index initiator due to Close call")
		}
	}
}

func (initiator *Initiator) ClearIndex(ctx context.Context) error {
	initiator.log.Info("clearing index")
	initiator.mu.Lock()
	defer initiator.mu.Unlock()
	initiator.indexedComics = make(map[string][]string)
	return nil
}

func (initiator *Initiator) Close() error {
	close(initiator.stopCh)
	return nil
}
