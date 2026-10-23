// Package settings serves the settings page as a local web page.
//
// The page is only reachable from this machine (127.0.0.1) and only with the
// random token contained in the URL opened by the application. Changes are
// sent to the application as ApplySettingsCommand, which saves and applies
// them; reading (bridge lights, location default) is done directly.
package settings

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sync"
	"time"

	"com.github.yveskaufmann/hue-lighter/internal/config"
	"com.github.yveskaufmann/hue-lighter/internal/geo"
	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	"com.github.yveskaufmann/hue-lighter/internal/messages"
	log "github.com/sirupsen/logrus"
)

//go:embed assets
var assets embed.FS

const sessionCookie = "hue_lighter_session"

// LightCatalog lists the lights known to the bridge.
type LightCatalog interface {
	GetAllLights() (*hueclient.LightList, error)
}

// LocationProvider determines a default location, see geo.Default.
type LocationProvider func(ctx context.Context, logger *log.Entry) (geo.Location, error)

// Server serves the settings page on a random local port.
type Server struct {
	logger   *log.Entry
	lights   LightCatalog
	locate   LocationProvider
	commands chan<- messages.Command

	mu         sync.Mutex
	config     *config.Config
	httpServer *http.Server
	host       string
	token      string
}

func NewServer(lights LightCatalog, locate LocationProvider, commands chan<- messages.Command, logger *log.Entry) *Server {
	return &Server{
		logger:   logger.WithField("component", "SettingsServer"),
		lights:   lights,
		locate:   locate,
		commands: commands,
	}
}

// SetConfig updates the config shown on the settings page. The server keeps
// its own copy, so it never shares state with the application.
func (s *Server) SetConfig(cfg *config.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg.Clone()
}

func (s *Server) currentConfig() *config.Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config.Clone()
}

// Start starts the server unless it is already running and returns the URL
// (including the access token) to open in the browser.
func (s *Server) Start() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.httpServer == nil {
		token, err := newToken()
		if err != nil {
			return "", err
		}

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return "", fmt.Errorf("failed to listen for settings page: %w", err)
		}

		s.token = token
		s.host = listener.Addr().String()
		s.httpServer = &http.Server{
			Handler:           s.routes(),
			ReadHeaderTimeout: 10 * time.Second,
		}

		go func(server *http.Server) {
			if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.logger.WithError(err).Error("Settings page server stopped unexpectedly")
			}
		}(s.httpServer)
		s.logger.Infof("Settings page available at http://%s", s.host)
	}

	return fmt.Sprintf("http://%s/?token=%s", s.host, s.token), nil
}

// Stop shuts the server down if it is running.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	server := s.httpServer
	s.httpServer = nil
	s.mu.Unlock()

	if server == nil {
		return nil
	}
	s.logger.Info("Stopping settings page server")
	return server.Shutdown(ctx)
}

func (s *Server) routes() http.Handler {
	static, _ := fs.Sub(assets, "assets")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex(static))
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))
	mux.HandleFunc("GET /api/settings", s.handleGetSettings)
	mux.HandleFunc("PUT /api/settings", s.handlePutSettings)
	mux.HandleFunc("GET /api/bridge/lights", s.handleGetBridgeLights)
	mux.HandleFunc("GET /api/location/default", s.handleGetLocationDefault)
	return s.protect(mux)
}

// protect only lets requests through which are sent to our own address (no
// DNS rebinding) and carry the session token, either in the URL opened by the
// application or in the session cookie set in exchange for it.
func (s *Server) protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		host, token := s.host, s.token
		s.mu.Unlock()

		if r.Host != host {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if queryToken := r.URL.Query().Get("token"); queryToken != "" && r.URL.Path == "/" {
			if !tokensEqual(queryToken, token) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			// Exchange the token for a cookie and drop it from the address bar.
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookie,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		cookie, err := r.Cookie(sessionCookie)
		if err != nil || !tokensEqual(cookie.Value, token) {
			http.Error(w, "forbidden: please open the settings via the tray icon", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleIndex(static fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFileFS(w, r, static, "index.html")
	}
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to create settings page token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func tokensEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
