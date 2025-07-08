package model

type Setting struct {
	ID           string
	Name         string
	Type         string // BOOLEAN | STRING | INTEGER
	DefaultValue string
}
