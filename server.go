// Package statuspages provide tools to implement status pages on your services.
package statuspages

import (
	"context"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/http/pprof"
	runtime_pprof "runtime/pprof"
	"strings"
	"sync"
)

// Service handles requests to /status pages.
type Service interface {
	// MainPageEntry returns a HTML snippet to include on the main page.
	MainPageEntry(name string) (string, error)

	// ServicePage handle requests to this service.
	ServicePage(name string, w http.ResponseWriter, r *http.Request) error
}

// Server (statuspages.Server) is a status pages HTTP server.
type Server struct {
	server       *http.Server
	listener     net.Listener
	bindRootPath bool

	mutex           sync.RWMutex // Protects everything below.
	names           map[Service]string
	services        map[string]Service
	orderedServices []Service
	httpMux         http.ServeMux
}

type Option func(s *Server)

// NewServer creates a new server with provided options.
func NewServer(opts ...Option) *Server {
	s := &Server{
		names:    make(map[Service]string),
		services: make(map[string]Service),
	}
	for _, o := range opts {
		o(s)
	}
	s.AddService("Base", newBaseService())
	return s
}

// WithHttpServer allows users to provide their own http server.
func WithHttpServer(server *http.Server) Option {
	return func(s *Server) {
		s.server = server
	}
}

// WithListener allows users to provide their own listener.
func WithListener(listener net.Listener) Option {
	return func(s *Server) {
		s.listener = listener
	}
}

// WithBindRootPath will bind the root path "/" to the status pages. This is useful when the
// web server is dedicated to status pages, and should be avoided if the web server is also used
// for another purpose.
// The /status path is always bound anyway.
func WithBindRootPath() Option {
	return func(s *Server) {
		s.bindRootPath = true
	}
}

// Add a new service on the status page.
// If another service with this name alreay exists, a suffix will be appended to the service name.
// The possibly suffixed service name is returned, then passed back to each Service.MainPageEntry
// and Service.ServicePage calls.
// If the service is already registred, this has no effect and returns the current name.
func (s *Server) AddService(name string, service Service) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if previousName := s.names[service]; previousName != "" {
		return previousName
	}
	return s.addService(name, service)
}

// AddService with mutex already locked.
func (s *Server) addService(name string, service Service) string {
	if name == "" {
		name = "NoName"
	}
	suffixedName := name
	if s.services[suffixedName] != nil {
		for i := 2; ; i++ {
			suffixedName = name + fmt.Sprint(i)
			if s.services[suffixedName] == nil {
				break
			}
		}
	}
	s.services[suffixedName] = service
	s.names[service] = suffixedName
	s.orderedServices = append(s.orderedServices, service)
	return suffixedName
}

// RemoveService removes a service from the status page. This has no effect if the service
// is not registered.
func (s *Server) RemoveService(service Service) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	name := s.names[service]
	if name == "" {
		return
	}
	delete(s.services, name)
	delete(s.names, service)
	for i, item := range s.orderedServices {
		if item == service {
			s.orderedServices = append(s.orderedServices[:i], s.orderedServices[i+1:]...)
			return
		}
	}
}

// AddHandler adds a http handler to the status pages web server.
func (s *Server) AddHandler(path string, handler http.Handler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.httpMux.Handle(path, handler)
}

// Run runs the server until context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	if s.server == nil {
		s.server = new(http.Server)
	}
	if s.listener == nil {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return err
		}
		s.listener = listener
	}

	s.server.Handler = &s.httpMux

	if s.bindRootPath {
		s.httpMux.Handle("/", s)
	}
	s.httpMux.Handle("/status", s)
	s.httpMux.HandleFunc("/pprof/", pprof.Index)
	s.httpMux.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	s.httpMux.HandleFunc("/pprof/profile", pprof.Profile)
	s.httpMux.HandleFunc("/pprof/symbol", pprof.Symbol)
	s.httpMux.HandleFunc("/pprof/trace", pprof.Trace)

	for _, profile := range runtime_pprof.Profiles() {
		s.httpMux.Handle("/pprof/"+profile.Name(), pprof.Handler(profile.Name()))
	}

	go func() {
		<-ctx.Done()
		s.server.Close() // nolint: errcheck
	}()

	return s.server.Serve(s.listener)
}

// ServeHTTP implements http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var err error

	defer func() {
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(html.EscapeString(err.Error()))) // nolint:errcheck
		}
	}()

	serviceName := ""
	switch r.Method {
	case "GET":
		if params, ok := r.URL.Query()["service"]; ok {
			serviceName = params[0]
		}
	case "POST":
		if err = r.ParseForm(); err == nil {
			serviceName = r.FormValue("service")
		}
	}

	var service Service
	if serviceName != "" {
		service = s.services[serviceName]
		if service == nil {
			err = fmt.Errorf("unknown service name %q", serviceName)
		} else {
			err = service.ServicePage(serviceName, w, r)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var sb strings.Builder

	err = Begin.Execute(&sb, "Main")
	if err != nil {
		return
	}

	for _, service = range s.orderedServices {
		name := s.names[service]
		serviceEntry.Execute(&sb, name)
		var html string
		html, err = service.MainPageEntry(name)
		if err != nil {
			return
		}
		_, err = sb.WriteString(html)
		if err != nil {
			return
		}
	}
	_, err = sb.WriteString(End)
	if err != nil {
		return
	}
	_, err = w.Write([]byte(sb.String()))
	if err != nil {
		// Not much we can do here, write to the client failed, probably a network issue.
		// Just reset err to make the defer() noop.
		err = nil
	}
}
