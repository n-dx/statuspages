package statuspages

import (
	"net/http"
	"os"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
)

type baseService struct {
	startTime time.Time
}

type mainPageTemplateParams struct {
	ProcessName  string
	Now          time.Time
	StartTime    time.Time
	StartTimeAgo string
}

type servicePageTemplateParams struct {
	CommandLine []string
	Env         []string
}

func newBaseService() *baseService {
	return &baseService{startTime: time.Now()}
}

// MainPageEntry implements Service interface.
func (b *baseService) MainPageEntry(name string) (string, error) {
	var sb strings.Builder
	now := time.Now()

	mainPageTemplate.Execute(&sb, &mainPageTemplateParams{
		ProcessName:  os.Args[0],
		Now:          now,
		StartTime:    b.startTime,
		StartTimeAgo: humanize.RelTime(b.startTime, now, "ago", "in the future"),
	})
	return sb.String(), nil
}

// ServicePage implements Service interface.
func (b *baseService) ServicePage(name string, w http.ResponseWriter, r *http.Request) error {
	var sb strings.Builder
	if err := Begin.Execute(&sb, name); err != nil {
		return err
	}
	if err := servicePageTemplate.Execute(&sb, &servicePageTemplateParams{
		CommandLine: os.Args,
		Env:         os.Environ(),
	}); err != nil {
		return err
	}
	if _, err := sb.WriteString(End); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte(sb.String()))
	return err
}
