package models

import "time"

type User struct {
	ID        int       `db:"id,primarykey,autoincrement"`
	Name      string    `db:"name,size:255" validate:"required,min=3"`
	Email     string    `db:"email,unique" validate:"required,email"`
	Age       int       `db:"age" validate:"min=18"`
	CreatedAt time.Time `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type Post struct {
	ID      int    `db:"id,primarykey,autoincrement"`
	Title   string `db:"title,size:255" validate:"required"`
	Content string `db:"content"`
	UserID  int    `db:"user_id"`
	User    *User  `db:"-"`
} 