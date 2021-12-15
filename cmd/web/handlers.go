package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"erik.web/pkg/models"
)

// обработчик для маршрута /
func (app *application) home(w http.ResponseWriter, r *http.Request) {

	// проверка url так как / обрабатывает все несуществующеи маршруты.
	// Отправляем ошибку 404 и завершаем функцию
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	s, err := app.snippets.Lastest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, r, "home.page.html", &templateData{
		Snippets: s,
	})
}

// обработчик для маршрута /sippets
func (app *application) showSnippets(w http.ResponseWriter, r *http.Request) {
	// Получаем значение параметра ?id из URL
	// Пытаемся полученную строку приобразовать в число.
	// если есть ошибка или полученное число отрицательное то
	// возвращаем ошибку 404
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	// Получение записи из таблицы базы данных по id.
	s, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, r, "show.page.html", &templateData{
		Snippet: s,
	})

}
func (app *application) newSnippet(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.html", &templateData{})
}

// обработчик для маршрута /snippet/create
func (app *application) createSnippets(w http.ResponseWriter, r *http.Request) {
	// проверка http метода запроса от клиента если не POST то вернуть ошибку 405
	// и указать не обходимый метод, после завершить работу функции.
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод запрещен!", http.StatusMethodNotAllowed)
		app.clientError(w, http.StatusMethodNotAllowed)
		return

	}

	id, err := app.snippets.Insert(
		r.PostFormValue("title"),
		r.PostFormValue("snippet"),
		r.PostFormValue("time"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", id), http.StatusSeeOther)

}
