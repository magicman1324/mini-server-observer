package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/pingan/hydra/internal/util"
	"github.com/pingan/hydra/internal/model"
)

// RuleRepository defines the interface for alert rule storage.
type RuleRepository interface {
	Create(rule *model.Rule) error
	Update(rule *model.Rule) error
	Delete(id string) error
	GetByID(id string) (*model.Rule, error)
	List(page, pageSize int) ([]*model.Rule, int, error)
	ListEnabled() ([]*model.Rule, error)
}

// MemRuleRepo is an in-memory implementation for development/testing.
type MemRuleRepo struct {
	mu    sync.RWMutex
	rules map[string]*model.Rule
}

func NewMemRuleRepo() *MemRuleRepo {
	return &MemRuleRepo{rules: make(map[string]*model.Rule)}
}

func (r *MemRuleRepo) Create(rule *model.Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rule.ID = util.NewID()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	if !rule.Enabled {
		rule.Enabled = true
	}
	r.rules[rule.ID] = rule
	return nil
}

func (r *MemRuleRepo) Update(rule *model.Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.rules[rule.ID]
	if !ok {
		return fmt.Errorf("rule %s not found", rule.ID)
	}
	rule.CreatedAt = existing.CreatedAt
	rule.UpdatedAt = time.Now()
	r.rules[rule.ID] = rule
	return nil
}

func (r *MemRuleRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rules, id)
	return nil
}

func (r *MemRuleRepo) GetByID(id string) (*model.Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rule, ok := r.rules[id]
	if !ok {
		return nil, fmt.Errorf("rule %s not found", id)
	}
	return rule, nil
}

func (r *MemRuleRepo) List(page, pageSize int) ([]*model.Rule, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*model.Rule
	for _, rule := range r.rules {
		all = append(all, rule)
	}

	total := len(all)
	start := (page - 1) * pageSize
	if start >= total {
		return []*model.Rule{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (r *MemRuleRepo) ListEnabled() ([]*model.Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var enabled []*model.Rule
	for _, rule := range r.rules {
		if rule.Enabled {
			enabled = append(enabled, rule)
		}
	}
	return enabled, nil
}
