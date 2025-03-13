package paginator2

import (
	"sync"
	"time"
)

// Manager is the main controller for the paginator. It contains the
// configuration used by all paginators, as well as all active
// paginators
type Manager struct {
	mutex      sync.Mutex
	paginators map[string]*Paginator
}

// NewManager creates a new paginator manager. It takes a variadic number of ConfigOpt
// options to configure the paginator. The default configuration is used if no
// options are provided.  The manager starts a goroutine to clean up expired paginators.
// The cleanup goroutine runs every 30 seconds and removes expired paginators.
func NewManager(opts ...ConfigOpt) *Manager {
	config := &defaultConfig
	config.Apply(opts)
	manager := &Manager{
		paginators: map[string]*Paginator{},
	}
	manager.startCleanup()

	return manager
}

// add adds a paginator to the manager.
func (m *Manager) Add(paginator *Paginator) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.paginators[paginator.id] = paginator
}

// remove removes a paginator from the manager.
func (m *Manager) Remove(paginatorID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.remove(paginatorID)
}

// remove removes a paginator from the manager and performs any necessary cleanup.
// It contains the shared logic used by `Remove“ and `cleanup`.
func (m *Manager) remove(paginatorID string) {
	// TODO: remove components?
	delete(m.paginators, paginatorID)
}

// cleanup removes expired paginators from the manager.
func (m *Manager) cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for _, p := range m.paginators {
		if !p.cxpiry.IsZero() && p.cxpiry.After(now) {
			m.remove(p.id)
		}
	}
}

// startCleanup starts a goroutine that cleans up expired paginators every minute.
func (m *Manager) startCleanup() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			m.cleanup()
		}
	}()
}
