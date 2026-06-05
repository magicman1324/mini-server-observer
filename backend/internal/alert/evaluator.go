package alert

import (
	"log"
	"sync"
	"time"

	"github.com/pingan/hydra/internal/model"
	"github.com/pingan/hydra/internal/repository"
)

// Evaluator evaluates incoming metric points against alert rules.
type Evaluator struct {
	mu         sync.RWMutex
	ruleRepo   repository.RuleRepository
	alertRepo  repository.AlertRepository
	lifecycle  *LifecycleManager
	dedup      *DedupTracker
	alertCh    chan<- *model.Alert // notification to dispatcher (PR5)
}

// NewEvaluator creates an alert evaluator.
func NewEvaluator(ruleRepo repository.RuleRepository, alertRepo repository.AlertRepository, notifyCh chan<- *model.Alert) *Evaluator {
	return &Evaluator{
		ruleRepo:  ruleRepo,
		alertRepo: alertRepo,
		lifecycle: NewLifecycleManager(),
		dedup:     NewDedupTracker(60 * time.Second),
		alertCh:   notifyCh,
	}
}

// ProcessPoints evaluates a batch of metric points against all enabled rules.
func (e *Evaluator) ProcessPoints(points []*model.MetricPoint) {
	if len(points) == 0 {
		return
	}

	e.mu.RLock()
	rules, err := e.ruleRepo.ListEnabled()
	e.mu.RUnlock()

	if err != nil {
		log.Printf("alert evaluator: failed to load rules: %v", err)
		return
	}

	now := time.Now()

	for _, point := range points {
		for _, rule := range rules {
			if rule.Metric != point.Name {
				continue
			}

			state, shouldEmit, msg := e.lifecycle.Evaluate(rule, point.Host, point.Region, point.Value, now)

			if shouldEmit {
				if state == StateFiring {
					if !e.dedup.ShouldFire(rule.ID, point.Host) {
						continue
					}

					alert := &model.Alert{
						RuleID:    rule.ID,
						RuleName:  rule.Name,
						Host:      point.Host,
						Region:    point.Region,
						Severity:  rule.Severity,
						Metric:    point.Name,
						Value:     point.Value,
						Threshold: rule.Threshold,
						Message:   msg,
						Status:    string(model.AlertFiring),
						FiredAt:   now,
						Labels:    rule.Labels,
					}

					if err := e.alertRepo.Create(alert); err != nil {
						log.Printf("alert evaluator: failed to create alert: %v", err)
						continue
					}

					log.Printf("ALERT FIRING: %s | %s: %s=%.2f (threshold: %.2f)",
						rule.Name, point.Host, point.Name, point.Value, rule.Threshold)

					if e.alertCh != nil {
						select {
						case e.alertCh <- alert:
						default:
							log.Printf("alert evaluator: notification channel full")
						}
					}
				} else {
					// Resolution message — log for now, notification in PR5
					log.Printf("ALERT RESOLVED: %s | %s: %s=%.2f back to normal",
						rule.Name, point.Host, point.Name, point.Value)
				}
			}
		}
	}
}

// ReloadRules refreshes the evaluator's rule cache (can be called on rule CRUD).
func (e *Evaluator) ReloadRules() {
	// Rules are loaded from repo on each ProcessPoints call,
	// so no explicit cache refresh is needed for the in-memory repo.
}
