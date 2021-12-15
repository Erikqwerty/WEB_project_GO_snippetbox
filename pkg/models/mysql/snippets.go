package mysql

import (
	"database/sql"
	"errors"

	"erik.web/pkg/models"
)

// Определение типа который обертывает пул подключений к базе данных.
// Чтобы не писать методы для sql.DB стандартной библиотеке.
// Также использование такой структуры с методами позволит создать интерфейс.
// Для написание функций тестирования.
type SnippetModel struct {
	DB *sql.DB
}

// Insert - Метод для создания заметок в базе данных.
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	// текст sql запроса для записи данных.
	stmt := `INSERT INTO snippets (title, content, created, expires) VALUES
	(?,?,
	UTC_TIMESTAMP(),
	DATE_ADD(UTC_TIMESTAMP(),
	INTERVAL ? DAY))`
	// Вызываем в встроеном пуле подключений метод Exec()
	// В метод передаем сам запрос и переменные.
	// Метод возвращает в переменную result некоторый ответ на запрос
	// записи данных в бд. В обьекте sql.Result.
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}
	// Используем метод LastInsertId чтобы получить последний ID созданной записи
	// В таблицу snippets.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}

	// преобразуем id тип int так как изначально id имеет тип int64
	return int(id), nil
}

// Get - Метод для получения заметки из базы данных по id.
func (m *SnippetModel) Get(id int) (*models.Snippets, error) {
	//  SQL запрос для получения данных из таблицы по ID заметки.
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`

	// Выполняем SQL запрос для метода QueryRow. Получаем данные записи.
	row := m.DB.QueryRow(stmt, id)

	// указатель на структуру Snippets.
	s := &models.Snippets{}
	// С помощью метода Scan записываем полученные данные в структуру Snippets.
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// если ошибка обнаружена то возвращаем ошибку из models.ErrNoRecord.
		// Подходящей записи не найдено.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

// Lastest - Метод для получения 10 последних заметок.
func (m *SnippetModel) Lastest() ([]*models.Snippets, error) {
	// Пишем SQL запрос, который мы хотим выполнить.
	stmt := `SELECT id, title, content, created, expires FROM snippets
WHERE expires > UTC_TIMESTAMP() ORDER BY created DESC LIMIT 10`

	// Используем метод Query() для выполнения нашего SQL запроса.
	// В ответ мы получим sql.Rows, который содержит результат нашего запроса.
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// Инициализируем пустой срез для хранения объектов models.Snippets.
	var snippets []*models.Snippets

	// Используем rows.Next() для перебора результата. Этот метод предоставляем
	// первый а затем каждую следующею запись из базы данных для обработки
	// методом rows.Scan().
	for rows.Next() {
		// Создаем указатель на новую структуру Snippet
		s := &models.Snippets{}
		// Используем rows.Scan(), чтобы скопировать значения полей в структуру.
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		// Добавляем структуру в срез.
		snippets = append(snippets, s)
	}

	// Когда цикл rows.Next() завершается, вызываем метод rows.Err(), чтобы узнать
	// если в ходе работы какая либо ошибка.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
