package store

import (
	"testing"
)

// TestStore ...
func TestStore(t *testing.T, tables ...string) *Store {
	t.Helper() // показывает это функция - хелпер». Если тест упадёт, в выводе будет строка вызова из теста, а не из хелпера

	config := NewConfig()
	s := New(config)
	if err := s.Open(); err != nil {
		t.Fatal(err)
	}

	// Очистка ПЕРЕД тестом
	for _, table := range tables {
		if _, err := s.db.Exec("TRUNCATE TABLE " + table); err != nil {
			t.Fatal(err)
		}
	}

	// Закрытие ПОСЛЕ теста
	t.Cleanup(func() {
		s.Close()
	})

	return s
}
