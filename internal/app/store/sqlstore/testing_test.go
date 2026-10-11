package sqlstore

import (
	"database/sql"
	"testing"

	"github.com/akidamura/http-rest-api/internal/app/store"
	_ "github.com/go-sql-driver/mysql"
)

// NewTestStore создаёт Store для интеграционных тестов.
// Подключается к тестовой БД, очищает указанные таблицы
// перед тестом и закрывает соединение после через t.Cleanup.
func NewTestStore(t *testing.T, tables ...string) *Store {
	t.Helper() // показывает это функция - хелпер». Если тест упадёт, в выводе будет строка вызова из теста, а не из хелпера

	db, err := sql.Open("mysql", store.BuildDSN())
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	// Очистка ПЕРЕД тестом
	for _, table := range tables {
		if _, err := db.Exec("TRUNCATE TABLE " + table); err != nil {
			t.Fatal(err)
		}
	}

	// Закрытие ПОСЛЕ теста
	t.Cleanup(func() {
		db.Close()
	})

	return New(db)
}
