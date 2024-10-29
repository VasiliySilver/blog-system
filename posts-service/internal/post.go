// posts-service/internal/models/post.go
package models

import "time"

type Post struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    AuthorID  string    `json:"author_id"`
   CreatedAt time.Time `json:"created_at"`
}
