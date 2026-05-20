package sqlutils

import (
	"database/sql"
	"database/sql/driver"
)

// IsZero проверяет, является ли заданное значение нулевым.
// Возвращает true, если значение является нулевым, в противном случае - false.
func IsZero[T comparable](v T) bool {
	return v == *new(T)
}

type SqlNull interface {
	sql.Scanner
	driver.Valuer
}

// SetOrNull устанавливает значение m равным v, если v не zero value.
func SetOrNull[T comparable](m SqlNull, v T) {
	if !IsZero(v) {
		_ = m.Scan(v)
	}
}
