package storage

import (
	"errors"

	"github.com/google/uuid"
)

type MemoryStorage struct {
	rules []Rule
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		rules: []Rule{
			{ID: uuid.New(), Name: "Test Feature", Backend: "http://localhost:8080", Percent: 50},
			{ID: uuid.New(), Name: "New Design", Backend: "http://localhost:8081", Percent: 50},
		},
	}
}

func (ms *MemoryStorage) GetAll() ([]Rule, error) {
	return ms.rules, nil
}

func (ms *MemoryStorage) GetByID(id uuid.UUID) (*Rule, error) {
	for i, rule := range ms.rules {
		if rule.ID == id {
			return &ms.rules[i], nil
		}
	}

	return nil, NotFoundErr
}

func (ms *MemoryStorage) Add(rule Rule) error {
	for _, r := range ms.rules {
		if r.ID == rule.ID {
			return errors.New("rule with this ID already exists")
		}
	}

	ms.rules = append(ms.rules, rule)
	return nil
}

func (ms *MemoryStorage) Update(rule Rule) error {
	for i := range ms.rules {
		if ms.rules[i].ID == rule.ID {
			ms.rules[i].Name = rule.Name
			ms.rules[i].Percent = rule.Percent
			return nil
		}
	}

	return NotFoundErr
}

func (ms *MemoryStorage) Delete(id uuid.UUID) error {
	for i, rule := range ms.rules {
		if rule.ID == id {
			ms.rules = append(ms.rules[:i], ms.rules[i+1:]...)
			return nil
		}
	}

	return NotFoundErr
}
