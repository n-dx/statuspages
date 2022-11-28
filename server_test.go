package statuspages

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddService(t *testing.T) {
	s := NewServer()

	assert.Equal(t, "name", s.AddService("name", &MockService{}))
	assert.Equal(t, "name2", s.AddService("name", &MockService{}))
}

func TestRemoveService(t *testing.T) {
	s := NewServer()

	s1 := &MockService{}
	s.AddService("one", s1)
	s2 := &MockService{}
	s.AddService("two", s2)

	s.RemoveService(s1)

	assert.Nil(t, s.services["one"])
	assert.Equal(t, "", s.names[s1])
	assert.Equal(t, s.orderedServices[1], s2) // First entry is Base service.

}

func TestRun(t *testing.T) {
	s := NewServer()
	s.AddService("x<y", &MockService{})
	ctx, cancelFunc := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	for func() bool {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		return s.listener == nil
	}() {
	}

	addrPort := s.listener.Addr().String()

	for retry := 0; ; retry++ {
		resp, err := http.Get("http://" + addrPort + "/status")
		if err != nil && retry < 10 {
			retry++
			time.Sleep(10 * time.Millisecond)
			continue
		}
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Contains(t, string(body), "<title>Main status page</title>")
		// Check correct escape of "x<y" un URL and in displayed text.
		assert.Contains(t, string(body), "<a href=\"./status?service=x%3cy\">x&lt;y</a></h1>")
		break
	}

	cancelFunc()
	<-done
}

// MockService implements Service interface.
type MockService struct {
	name string
}

// MainPageEntry implements Service interface.
func (m *MockService) MainPageEntry(name string) (string, error) {
	m.name = name
	return name + " service", nil
}

// ServicePage implements Service interface.
func (m *MockService) ServicePage(name string, w http.ResponseWriter, r *http.Request) error {
	_, err := w.Write([]byte(name + " service status page"))
	return err
}
