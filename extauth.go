package extauth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/rs/zerolog/log"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("extauth", parseCaddyfile)
}

// Middleware implements an HTTP handler that calls an external service for authentication
type Middleware struct {
	Endpoint            string            // http endpoint endpoint for the auth request
	Timeout             time.Duration     // timeout for the auth request
	CopyRequestHeaders  []string          // headers to copy from the incoming request to the auth request
	CopyResponseHeaders []string          // headers to copy from the auth response to the incoming request
	SetHeaders          map[string]string // headers to set in the auth request
	httpClient          *http.Client
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.extauth",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Middleware) Provision(ctx caddy.Context) error {
	m.httpClient = &http.Client{
		Timeout: m.Timeout,
	}
	return nil
}

// Validate implements caddy.Validator.
func (m *Middleware) Validate() error {
	if m.Endpoint == "" {
		return errors.New("'endpoint' is required")
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// create auth request
	authReq, err := http.NewRequest(http.MethodGet, m.Endpoint, nil)
	if err != nil {
		return err
	}

	// copy headers from the incoming request
	for _, name := range m.CopyRequestHeaders {
		authReq.Header.Set(name, r.Header.Get(name))
	}

	// set additional headers
	for key, val := range m.SetHeaders {
		if val == "{http.request.uri}" {
			val = r.URL.RequestURI()
		}
		if val == "{http.request.method}" {
			val = r.Method
		}
		authReq.Header.Set(key, val)
	}

	// perform the request
	resp, err := m.httpClient.Do(authReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		// something went wrong or the server responded with something != 200 -> reject request with 401
		log.Error().Str("url", r.URL.RequestURI()).Msg("failed to authenticate")
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	log.Info().Str("url", r.URL.RequestURI()).Msg("successfully authenticated")

	// copy the user defined response headers to the incoming request before going on to the next handler in the chain
	for _, name := range m.CopyResponseHeaders {
		val := resp.Header.Get(name)
		if val != "" {
			r.Header.Set(name, resp.Header.Get(name))
		}
	}

	// call next handler
	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) (err error) {
	m.SetHeaders = make(map[string]string)
	d.NextArg()
	for d.NextBlock(0) {
		switch d.Val() {
		case "endpoint":
			if !d.AllArgs(&m.Endpoint) {
				return d.ArgErr()
			}
		case "timeout":
			timeoutStr := ""
			if !d.AllArgs(&timeoutStr) {
				return d.ArgErr()
			}
			m.Timeout, err = time.ParseDuration(timeoutStr)
			if err != nil {
				return fmt.Errorf("can't parse timeout: %w", err)
			}
		case "copy-request-header":
			name := ""
			for d.Args(&name) {
				m.CopyRequestHeaders = append(m.CopyRequestHeaders, name)
			}
		case "copy-response-header":
			name := ""
			for d.Args(&name) {
				m.CopyResponseHeaders = append(m.CopyResponseHeaders, name)
			}
		case "set-header":
			var key, val string
			for d.AllArgs(&key, &val) {
				m.SetHeaders[key] = val
			}
		default:
			return d.Errf("%s not a valid extauth option", d.Val())
		}
	}
	if m.Timeout == 0 {
		m.Timeout = 1 * time.Second
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
