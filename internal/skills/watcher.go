package skills

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher   *fsnotify.Watcher
	mu        sync.Mutex
	onChanged func()
	debounce  time.Duration
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewWatcher(onChanged func()) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Watcher{
		watcher:   w,
		onChanged: onChanged,
		debounce:  500 * time.Millisecond,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// AddDir watches a skill directory recursively.
// Only directories that exist are watched.
func (w *Watcher) AddDir(dir string) {
	if _, err := os.Stat(dir); err != nil {
		return
	}
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			w.mu.Lock()
			_ = w.watcher.Add(path)
			w.mu.Unlock()
		}
		return nil
	})
}

// AddSkillDir watches a single skill directory (contains SKILL.md).
func (w *Watcher) AddSkillDir(dir string) {
	w.AddDir(dir)
}

// Start begins the event loop. Blocks until Stop() is called.
func (w *Watcher) Start() {
	var timer *time.Timer
	var debounceCh <-chan time.Time

	for {
		select {
		case <-w.ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if w.isSkillChange(event) {
				if timer != nil {
					timer.Stop()
				}
				timer = time.NewTimer(w.debounce)
				debounceCh = timer.C
			}
		case <-debounceCh:
			debounceCh = nil
			if w.onChanged != nil {
				w.onChanged()
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("skill watcher error: %v", err)
		}
	}
}

func (w *Watcher) isSkillChange(event fsnotify.Event) bool {
	name := filepath.Base(event.Name)
	if name != "SKILL.md" {
		// Also watch for new skill directories (Create on a dir)
		if event.Has(fsnotify.Create) {
			if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
				w.mu.Lock()
				_ = w.watcher.Add(event.Name)
				w.mu.Unlock()
			}
		}
		return false
	}
	return event.Has(fsnotify.Write) || event.Has(fsnotify.Create)
}

func (w *Watcher) Stop() {
	w.cancel()
	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.watcher.Close()
}
