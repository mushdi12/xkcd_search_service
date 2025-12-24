package core

type EventType string

const (
	EventTypeUpdating EventType = "update"
	EventTypeDropped  EventType = "drop"
)

type DBStats struct {
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
}

type Comics struct {
	ID    int
	URL   string
	Words []string
}
