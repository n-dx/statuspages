package statuspages

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseServiceMainPage(t *testing.T) {
	b := newBaseService()
	var err error
	b.startTime, err = time.Parse("02/01/2006", "01/01/2020")
	require.NoError(t, err)

	html, err := b.MainPageEntry("x<y")
	require.NoError(t, err)
	assert.Contains(t, html, "Start time: 00:00:00 01/01/2020")
	assert.Contains(t, html, "years ago")
}

func TestBaseServiceServicePage(t *testing.T) {
	b := newBaseService()

	w := httptest.NewRecorder()
	require.NoError(t, b.ServicePage("x<y", w, nil))

	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, "<title>x&lt;y status page</title>")
	assert.Contains(t, body, "/statuspages.test") // In command line.
	assert.Contains(t, body, "<li>PATH=")         // In environment.

}
