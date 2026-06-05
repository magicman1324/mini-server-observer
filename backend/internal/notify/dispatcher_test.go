package notify

import (
	"sync"
	"testing"
	"time"

	"github.com/pingan/hydra/internal/model"
)

// mockChannel records sent alerts for testing.
type mockChannel struct {
	mu     sync.Mutex
	name   string
	alerts []*model.Alert
}

func (m *mockChannel) Name() string            { return m.name }
func (m *mockChannel) Send(alert *model.Alert) error {
	m.mu.Lock()
	m.alerts = append(m.alerts, alert)
	m.mu.Unlock()
	return nil
}
func (m *mockChannel) Alerts() []*model.Alert {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*model.Alert, len(m.alerts))
	copy(result, m.alerts)
	return result
}

func TestDispatcher_Dispatch(t *testing.T) {
	ch := make(chan *model.Alert, 10)
	d := NewDispatcher(ch)

	mock := &mockChannel{name: "mock-dingtalk"}
	d.Register(mock)
	d.Start()
	<-d.Ready()

	// Send alert
	alert := &model.Alert{
		ID:       "alert-1",
		RuleName: "High CPU",
		Host:     "web-01",
		Severity: "critical",
		Metric:   "cpu_usage_percent",
		Value:    95.0,
		Status:   "firing",
	}
	ch <- alert
	close(ch)
	time.Sleep(50 * time.Millisecond)

	// Check the mock received it
	alerts := mock.Alerts()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].ID != "alert-1" {
		t.Errorf("alert ID = %q", alerts[0].ID)
	}
}

func TestDispatcher_MultipleChannels(t *testing.T) {
	ch := make(chan *model.Alert, 10)
	d := NewDispatcher(ch)

	mock1 := &mockChannel{name: "dingtalk"}
	mock2 := &mockChannel{name: "email"}
	d.Register(mock1)
	d.Register(mock2)
	d.Start()
	<-d.Ready()

	alert := &model.Alert{ID: "a-1", RuleName: "Test", Severity: "warning"}
	ch <- alert
	close(ch)
	time.Sleep(50 * time.Millisecond)

	if len(mock1.Alerts()) != 1 {
		t.Errorf("channel 1: expected 1 alert, got %d", len(mock1.Alerts()))
	}
	if len(mock2.Alerts()) != 1 {
		t.Errorf("channel 2: expected 1 alert, got %d", len(mock2.Alerts()))
	}
}

func TestRetryWithBackoff_Success(t *testing.T) {
	attempts := 0
	err := RetryWithBackoff("test", 3, func() error {
		attempts++
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}
