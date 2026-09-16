package cleaner_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/vyntechau/TelegramPublisher/internal/cleaner"
	"gopkg.in/telebot.v3"
)

type mockDeleter struct {
	mu           sync.Mutex
	deletedMsgs  []int
	deleteErrMap map[int]error
}

func newMockDeleter() *mockDeleter {
	return &mockDeleter{
		deletedMsgs:  make([]int, 0),
		deleteErrMap: make(map[int]error),
	}
}

func (m *mockDeleter) Delete(msg telebot.Editable) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	msgID, _ := msg.MessageSig()
	var id int
	if mObj, ok := msg.(*telebot.Message); ok {
		id = mObj.ID
	} else {
		_ = msgID
	}

	m.deletedMsgs = append(m.deletedMsgs, id)
	if err, exists := m.deleteErrMap[id]; exists {
		return err
	}
	return nil
}

func (m *mockDeleter) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.deletedMsgs)
}

func TestCleanerScheduleAndStop(t *testing.T) {
	mock := newMockDeleter()
	mock.deleteErrMap[102] = errors.New("message not found")

	c := cleaner.NewWithInterval(mock, 10*time.Millisecond)
	c.Start()

	// 1. Schedule multiple messages (one will succeed, one will return error)
	c.Schedule(123456, []int{101, 102}, 20*time.Millisecond)

	// 2. Schedule single message
	c.ScheduleSingle(123456, 103, 20*time.Millisecond)

	// 3. Schedule message in the far future (should not be purged yet)
	c.ScheduleSingle(123456, 999, 10*time.Second)

	// 4. Zero/Negative duration and empty slice branches
	c.Schedule(123456, []int{104}, 0)
	c.Schedule(123456, []int{105}, -1*time.Second)
	c.Schedule(123456, nil, 50*time.Millisecond)
	c.Schedule(123456, []int{}, 50*time.Millisecond)

	// Wait for the fast tasks to expire and ticker to trigger purgeExpired
	time.Sleep(100 * time.Millisecond)

	if mock.count() < 3 {
		t.Errorf("expected at least 3 messages processed, got %d", mock.count())
	}

	c.Stop()
}

func TestCleanerWithNilBot(t *testing.T) {
	c := cleaner.NewWithInterval(nil, 10*time.Millisecond)
	c.Start()

	c.ScheduleSingle(123456, 101, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	c.Stop()
}

func TestCleanerDefaultNewAndZeroInterval(t *testing.T) {
	c := cleaner.New(nil)
	c.Start()
	c.Stop()

	// Test zero/negative interval fallback in NewWithInterval
	c2 := cleaner.NewWithInterval(nil, 0)
	c2.Start()
	c2.Stop()
}
