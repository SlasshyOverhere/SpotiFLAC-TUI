package queue

import (
	"encoding/json"
	"sync"

	gobackend "github.com/zarz/spotiflac_android/go_backend"

	"spotiflac-cli/internal/backend"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "downloading"
	StatusDone      Status = "done"
	StatusError     Status = "error"
	StatusCancelled Status = "cancelled"
)

type Queued struct {
	Title      string
	RequestJSON string
}

type Item struct {
	Title    string
	ID       string
	Status   Status
	Progress backend.ProgressItem
	Result   *gobackend.DownloadResponse
	Error    string
}

type Manager struct {
	mu    sync.Mutex
	items []*Item
	byID  map[string]*Item
	ch    chan Queued
}

func New() *Manager {
	m := &Manager{
		byID: make(map[string]*Item),
		ch:   make(chan Queued, 64),
	}
	go m.worker()
	return m
}

func (m *Manager) Enqueue(items ...Queued) {
	for _, q := range items {
		m.mu.Lock()
		item := &Item{Title: q.Title, ID: q.RequestID(), Status: StatusQueued}
		m.items = append(m.items, item)
		m.byID[item.ID] = item
		m.mu.Unlock()
		m.ch <- q
	}
}

func (q Queued) RequestID() string {
	var req gobackend.DownloadRequest
	if err := json.Unmarshal([]byte(q.RequestJSON), &req); err == nil && req.ItemID != "" {
		return req.ItemID
	}
	return q.Title
}

func (m *Manager) Items() []Item {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Item, len(m.items))
	for i, it := range m.items {
		out[i] = *it
	}
	return out
}

func (m *Manager) Cancel(id string) {
	backend.CancelDownload(id)
	m.mu.Lock()
	if it, ok := m.byID[id]; ok && it.Status == StatusQueued {
		it.Status = StatusCancelled
	}
	m.mu.Unlock()
}

func (m *Manager) Refresh() {
	prog := backend.AllProgress()
	m.mu.Lock()
	for _, it := range m.items {
		if p, ok := prog[it.ID]; ok {
			it.Progress = p
			switch p.Status {
			case "downloading", "preparing", "finalizing":
				it.Status = StatusRunning
			case "completed":
				if it.Status != StatusError {
					it.Status = StatusDone
				}
			}
		}
	}
	m.mu.Unlock()
}

func (m *Manager) worker() {
	for q := range m.ch {
		m.mu.Lock()
		item := m.byID[q.RequestID()]
		if item == nil || item.Status == StatusCancelled {
			m.mu.Unlock()
			continue
		}
		item.Status = StatusRunning
		m.mu.Unlock()

		raw, err := backend.Download(q.RequestJSON)
		var resp gobackend.DownloadResponse
		if err == nil {
			err = json.Unmarshal([]byte(raw), &resp)
		}

		m.mu.Lock()
		if err != nil {
			item.Status = StatusError
			item.Error = err.Error()
		} else if item.Status == StatusCancelled {
			// leave cancelled
		} else if !resp.Success {
			item.Status = StatusError
			item.Error = resp.Error
			if resp.ErrorType == "cancelled" {
				item.Status = StatusCancelled
			}
		} else {
			item.Status = StatusDone
			item.Result = &resp
			item.Progress.Status = "completed"
		}
		m.mu.Unlock()
	}
}
