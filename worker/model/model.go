package model

type Message struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

type Response struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}
