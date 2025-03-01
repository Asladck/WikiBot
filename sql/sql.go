package sql

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type User struct {
	DiscordID int
	Nickname  string
	Data      string
	Language  string
}

func Connect() *pgx.Conn {
	connString := "postgres://postgres:12345678@localhost:5432/testBase"
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	fmt.Println("Успешное подключение к PostgreSQL!")
	return conn
}
