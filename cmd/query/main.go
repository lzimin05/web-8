package main

// некоторые импорты нужны для проверки
import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http" // пакет для поддержки HTTP протокола

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "lenya"
	password = "111222qqq"
	dbname   = "sandbox"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

func (dp *DatabaseProvider) SelectText() (string, error) {
	var msg string
	row := dp.db.QueryRow("SELECT message FROM query ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}
	return msg, nil
}

func (dp *DatabaseProvider) InsertText(msg string) error {
	_, err := dp.db.Exec("INSERT INTO query (message) VALUES ($1)", msg)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handlers) HandleApiUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		name, err := h.dbProvider.SelectText()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		w.Write([]byte("Hello," + name + "!"))
	}
	if r.Method == "POST" {
		name := r.URL.Query().Get("name")
		if name != "" {
			err := h.dbProvider.InsertText(name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
		}
	}
}

func main() {
	address := flag.String("address", "127.0.0.1:9000", "адрес для запуска сервера")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем провайдер для БД с набором методов
	dp := DatabaseProvider{db: db}
	// Создаем экземпляр структуры с набором обработчиков
	h := Handlers{dbProvider: dp}

	// Регистрируем обработчики
	http.HandleFunc("/api/user", h.HandleApiUser)

	// Запускаем веб-сервер на указанном адресе
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
