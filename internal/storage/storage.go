package storage

import (
	"errors"

	"github.com/google/uuid"
)

type Rule struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Percent int       `json:"percent"`
}

type Storage interface {
	GetAll() ([]Rule, error)
	GetByID(id uuid.UUID) (*Rule, error)
	Add(rule Rule) error
	Update(rule Rule) error
	Delete(id uuid.UUID) error
}

var NotFoundErr = errors.New("not found")
