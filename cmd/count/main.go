package main

// некоторые импорты нужны для проверки
import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http" // пакет для поддержки HTTP протокола
	"strconv"

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
	row := dp.db.QueryRow("SELECT count FROM count")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}
	return msg, nil
}

func (dp *DatabaseProvider) InsertText(msg int) error {
	_, err := dp.db.Exec("UPDATE count SET count = ($1)", msg)
	if err != nil {
		return err
	}

	return nil
}

func (dp *DatabaseProvider) GetCount() (int, error) {
	var msg string
	row := dp.db.QueryRow("SELECT count FROM count")
	err := row.Scan(&msg)
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(msg)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (h *Handlers) HandleCount(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		count, err := h.dbProvider.SelectText()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		w.Write([]byte("count = " + count))
	}
	if r.Method == "POST" {
		r.ParseForm()
		s := r.FormValue("count")
		if s == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("это не число"))

			return
		}
		number, err := strconv.Atoi(s)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("это не число"))

			return
		}
		count_old, err := h.dbProvider.GetCount()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		err = h.dbProvider.InsertText(number + count_old)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	}
}

func main() {
	address := flag.String("address", "127.0.0.1:3333", "адрес для запуска сервера")
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
	http.HandleFunc("/count", h.HandleCount)

	// Запускаем веб-сервер на указанном адресе
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
