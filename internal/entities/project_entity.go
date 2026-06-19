package entities

type Project struct {
	BaseEntity
	Name string `db:"name"`
}
