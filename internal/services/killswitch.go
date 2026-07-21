// internal/services/killswitch.go

package services

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Killswitch struct {
	shutdown chan struct{}
	once     sync.Once

	mu           sync.Mutex
	lastActivity time.Time

	timeout time.Duration
}

func NewKillswitch(timeoutMinutes int) *Killswitch {
	return &Killswitch{
		shutdown:     make(chan struct{}),
		lastActivity: time.Now().UTC(),
		timeout:      time.Duration(timeoutMinutes) * time.Minute,
	}
}

// TriggerShutdown closes the shutdown channel exactly once, used by both
// the inactivity monitor and the manual terminate listener.
func (k *Killswitch) TriggerShutdown() {
	k.once.Do(func() {
		close(k.shutdown)
		fmt.Println("+-----------------------------------------+")
		fmt.Println("| [KillSwitch] Shutdown triggered.        |")
		fmt.Println("+-----------------------------------------+")
	})
}

// WaitForShutdown returns a channel that closes when any shutdown path fires.
func (k *Killswitch) WaitForShutdown() <-chan struct{} {
	return k.shutdown
}

// ManualTerminateListener reads lines from a shared stdin channel (owned
// by main.go, since Ctrl+C's y/N prompt also needs stdin) and triggers
// shutdown on "terminate".
func (k *Killswitch) ManualTerminateListener(lines <-chan string) {
	for {
		select {
		case <-k.shutdown:
			return
		case line, ok := <-lines:
			if !ok {
				return
			}
			if strings.TrimSpace(strings.ToLower(line)) == "terminate" {
				fmt.Println("[KillSwitch] 'terminate' command received.")
				k.TriggerShutdown()
				return
			}
		}
	}
}

// UpdateActivity records the timestamp of the latest HTTP request.
func (k *Killswitch) UpdateActivity() {
	k.mu.Lock()
	k.lastActivity = time.Now().UTC()
	k.mu.Unlock()
}

// InactivityMonitor polls every 5s and triggers shutdown once the gap
// since the last activity exceeds the configured timeout.
func (k *Killswitch) InactivityMonitor() {
	fmt.Printf("[KillSwitch] Inactivity monitor started (timeout: %s).\n", k.timeout)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-k.shutdown:
			return
		case <-ticker.C:
			k.mu.Lock()
			elapsed := time.Since(k.lastActivity)
			k.mu.Unlock()

			if elapsed > k.timeout {
				fmt.Println("[KillSwitch] Inactivity timeout reached. Shutting down.")
				k.TriggerShutdown()
				return
			}
		}
	}
}
