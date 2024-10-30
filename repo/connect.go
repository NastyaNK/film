package repo

import (
	"fmt"
	"log"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Глобальная переменная для хранения соединения
var (
	DB   *sqlx.DB
	once sync.Once
)

// Repository структура для работы с базой данных
type Repository struct {
	db *sqlx.DB
}

func Connect() *sqlx.DB {
	once.Do(func() { // Код внутри этого блока выполнится только один раз
		var err error
		DB, err = sqlx.Connect("postgres", "host=localhost port=5432 user=anastasia password=2553 dbname=movie sslmode=disable")
		if err != nil {
			log.Fatal(err)
		}
	})
	return DB
}

func GetRepository() *Repository {
	conn := Connect()
	fmt.Println(conn)

	return &Repository{db: conn}
}
