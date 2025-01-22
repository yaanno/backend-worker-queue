package model

type BackendMessage struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

type WorkerMessage struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

type Message struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

type BackendResponse struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}
