package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// Помошник serverError создает сообщение ошибки после вызывает созданый логер и выводит текст ошибки.
// После возвращает ошибку http 500 "внутренняя ошибка сервера" в браузер.
// Используеться debug.Stack для возможности видеть полный путь к приложению через трассировку стека.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s \n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Помощник clientError будет возвращать ошибку пользователю по типу 400 "Не верный запрос".
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// Удобная оболочка вокруг clientError. Отправляет пользователю ошибку 404 "Страница не найдена".
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("шаблона не существует: %s", name))
		return
	}
	err := ts.Execute(w, td)
	if err != nil {
		app.serverError(w, err)
	}
}
