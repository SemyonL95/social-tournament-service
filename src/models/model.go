package models

type DB interface {
	Testdb()
}

type Model struct {
	DB
}

func InitModel(db DB) *Model {
	return &Model{
		 db,
	}
}

func (m *Model) test () {
	m.Testdb()
}