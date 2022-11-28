// Example status page. Displays a status page on localhost:8080.
package main

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/n-dx/statuspages"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	spServer := statuspages.NewServer(
		statuspages.WithBindRootPath(),
		statuspages.WithListener(listener))

	spServer.AddService("Example", new(ExampleService))
	spServer.Run(context.Background())
}

type ExampleService struct {
	name string
}

// MainPageEntry implements Service interface.
func (e *ExampleService) MainPageEntry(name string) (string, error) {
	e.name = name
	return "Example service", nil
}

// ServicePage implements Service interface.
func (e *ExampleService) ServicePage(name string, w http.ResponseWriter, r *http.Request) error {
	var sb strings.Builder
	if err := statuspages.Begin.Execute(&sb, name); err != nil {
		return err
	}

	// ...

	if _, err := sb.WriteString(statuspages.End); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte(sb.String()))
	return err
}
