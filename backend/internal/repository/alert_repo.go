package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/pingan/hydra/internal/util"
	"github.com/pingan/hydra/internal/model"
)

// AlertRepository defines the interface for alert storage.
type AlertRepository interface {
	Create(alert *model.Alert) error
	UpdateStatus(id, status string) error
	List(page, pageSize int, status, severity, host string) ([]*model.Alert, int, error)
	GetByID(id string) (*model.Alert, error)
	Stats() (firing int, resolved int)
}

// MemAlertRepo is an in-memory implementation for development/testing.
type MemAlertRepo struct {
	mu     sync.RWMutex
	alerts map[string]*model.Alert
}

func NewMemAlertRepo() *MemAlertRepo {
	return &MemAlertRepo{alerts: make(map[string]*model.Alert)}
}

func (r *MemAlertRepo) Create(alert *model.Alert) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	alert.ID = util.NewID()
	alert.FiredAt = time.Now()
	alert.Status = string(model.AlertFiring)
	r.alerts[alert.ID] = alert
	return nil
}

func (r *MemAlertRepo) UpdateStatus(id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	alert, ok := r.alerts[id]
	if !ok {
		return fmt.Errorf("alert %s not found", id)
	}
	alert.Status = status
	if status == string(model.AlertResolved) {
		now := time.Now()
		alert.ResolvedAt = &now
	}
	return nil
}

func (r *MemAlertRepo) List(page, pageSize int, status, severity, host string) ([]*model.Alert, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*model.Alert
	for _, a := range r.alerts {
		if status != "" && a.Status != status {
			continue
		}
		if severity != "" && a.Severity != severity {
			continue
		}
		if host != "" && a.Host != host {
			continue
		}
		filtered = append(filtered, a)
	}

	total := len(filtered)
	start := (page - 1) * pageSize
	if start >= total {
		return []*model.Alert{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return filtered[start:end], total, nil
}

func (r *MemAlertRepo) GetByID(id string) (*model.Alert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	alert, ok := r.alerts[id]
	if !ok {
		return nil, fmt.Errorf("alert %s not found", id)
	}
	return alert, nil
}

func (r *MemAlertRepo) Stats() (firing int, resolved int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.alerts {
		switch a.Status {
		case string(model.AlertFiring):
			firing++
		case string(model.AlertResolved):
			resolved++
		}
	}
	return
}
