package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"erik.web/pkg/models/mysql"
	_ "github.com/go-sql-driver/mysql"
)

// Тип структуры для зависимостей всего приложения.
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *mysql.SnippetModel
	templateCache map[string]*template.Template
}

// главная go рутина, точка входа программы
func main() {
	// Добавляем возможность запуска приложения с параметром.
	addr := flag.String("addr", ":4000", "Сетевой адресс http.")
	dsn := flag.String("dsn", "web:admin123@/snippetbox?parseTime=true", "Подключение к базе данных, пароль логин")
	// Извлекает введеные параметры из командной строки и записывает в указанные выше переменные.
	flag.Parse()

	// Создаем многоуровневые конкурентно безопасные логеры для ошибок и информации.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// вызываем функцию которая передает пул подключений к базе данных.
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	templateCache, err := NewtemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// создаем структуру зависимостей
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &mysql.SnippetModel{DB: db}, // записываем пул подключений в в модель SnippetModel.
		templateCache: templateCache,
	}

	// создаем структуру сервера на базе типа http.Server.
	// такой ход позволяет не прописывать параметры методу ListenAndServer
	// так как он будет брать эти пораметры из структуры
	// также передаем созданный логер  errorLog чтобы при возникновение ошибок сервер его использовал.
	// в Hanfler вызываем метод routes который верннет mux.
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	// вывод лога о запуске сервера и запуск сервера на 4000 порту, проверка ошибки.
	infoLog.Printf("запуск веб приложения на http://127.0.0.1%s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

// тип который содержит интерфейс описывающий метод Open
type neutFileSystem struct {
	fs http.FileSystem
}

// Вызываеться каждый раз когда http.FileServer получает запрос
// блокирут доступ к файловой системе для пользователя.
// Если пользователь пытаеться получить доступ к директории без файла index.html
// то вернеться ошибка 404, в случаях если этот файл есть то отдаст его.
//В остальных случаях выдаст запрашиваемый файл.
func (nfs neutFileSystem) Open(path string) (http.File, error) {
	// открываем вызываемый путь.
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// узнаем тип файла.
	s, err := f.Stat()
	if err != nil {
		fmt.Println(err.Error())
	}
	// Проверяем являеться ли вызываемый путь директорией.
	// Если это папка то проверяем есть ли в этой папке файл index.html
	// Если файл не существует возвращает ошибку os.ErrNotExist
	// Которая будет преобразованна через http.FileServer в ошибку 404.
	// Также вызываем метод f.Close() для открытого файла.
	// Таким образом блокируеться возможность лазить по файловой системе.
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		_, err := nfs.fs.Open(index)
		if err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, err
		}
	}
	// востальных случаях возвращем запрашиваемый файл.
	return f, nil
}

// Инициализует структуру пул подключения к базе данных.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, nil
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
