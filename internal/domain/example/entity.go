package example

type ExampleCounter struct {
	ID    int64 `db:"id" gorm:"primaryKey;autoIncrement:false"`
	Value int64 `db:"value"`
}

func (c *ExampleCounter) Increment(value int64) int64 {
	c.Value += value
	return c.Value
}
