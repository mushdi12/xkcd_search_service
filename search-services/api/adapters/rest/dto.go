package rest

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

type ComicsReply struct {
	Comics []Comics `json:"comics"`
	Total  int      `json:"total"`
}

type Comics struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type UpdateStats struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

type UpdateStatus struct {
	Status string `json:"status"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}
