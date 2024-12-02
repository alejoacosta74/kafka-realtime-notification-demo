package models

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Notification is a struct that represents a notification topic
type Notification struct {
	From    User   `json:"from"`
	To      User   `json:"to"`
	Message string `json:"message"`
}
