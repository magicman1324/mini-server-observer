package notify

import (
	"log"
	"sync"

	"github.com/pingan/hydra/internal/model"
)

// Dispatcher receives alert events and routes them to configured notification channels.
type Dispatcher struct {
	mu       sync.RWMutex
	channels []Channel
	alertCh  <-chan *model.Alert
	ready    chan struct{}
}

// Channel defines the interface for a notification channel (DingTalk, email, etc.).
type Channel interface {
	// Send delivers an alert notification.
	Send(alert *model.Alert) error
	// Name returns the channel identifier.
	Name() string
}

// NewDispatcher creates a dispatcher that reads from the given alert channel.
func NewDispatcher(alertCh <-chan *model.Alert) *Dispatcher {
	return &Dispatcher{
		alertCh: alertCh,
		ready:   make(chan struct{}),
	}
}

// Ready returns a channel that is closed when the dispatcher starts listening.
func (d *Dispatcher) Ready() <-chan struct{} {
	return d.ready
}

// Register adds a notification channel.
func (d *Dispatcher) Register(ch Channel) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.channels = append(d.channels, ch)
	log.Printf("notify: registered channel %q", ch.Name())
}

// Start begins listening for alert events and dispatching them.
func (d *Dispatcher) Start() {
	go func() {
		close(d.ready)
		for alert := range d.alertCh {
			d.mu.RLock()
			channels := make([]Channel, len(d.channels))
			copy(channels, d.channels)
			d.mu.RUnlock()

			for _, ch := range channels {
				go func(c Channel) {
					if err := c.Send(alert); err != nil {
						log.Printf("notify: channel %q failed to send alert %s: %v",
							c.Name(), alert.ID, err)
					}
				}(ch)
			}
		}
	}()
	log.Println("notify: dispatcher started")
}
