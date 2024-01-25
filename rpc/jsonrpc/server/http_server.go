// Commons for HTTP handling
package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/netutil"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

// Config is a RPC server configuration.
type Config struct {
	// see netutil.LimitListener
	MaxOpenConnections int
	// mirrors http.Server#ReadTimeout
	ReadTimeout time.Duration
	// mirrors http.Server#WriteTimeout
	WriteTimeout time.Duration
	// MaxBodyBytes controls the maximum number of bytes the
	// server will read parsing the request body.
	MaxBodyBytes int64
	// mirrors http.Server#MaxHeaderBytes
	MaxHeaderBytes int
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		MaxOpenConnections: 0, // unlimited
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		MaxBodyBytes:       int64(1000000), // 1MB
		MaxHeaderBytes:     1 << 20,        // same as the net/http default
	}
}

// Serve creates a http.Server and calls Serve with the given listener. It
// wraps handler with RecoverAndLogHandler and a handler, which limits the max
// body size to config.MaxBodyBytes.
//
// NOTE: This function blocks - you may want to call it in a go-routine.
func Serve(listener net.Listener, handler http.Handler, logger log.Logger, config *Config) error {
	logger.Info("serve", "msg", log.NewLazySprintf("Starting RPC HTTP server on %s", listener.Addr()))
	s := &http.Server{
		Handler:           RecoverAndLogHandler(maxBytesHandler{h: handler, n: config.MaxBodyBytes}, logger),
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
	}
	err := s.Serve(listener)
	logger.Info("RPC HTTP server stopped", "err", err)
	return err
}

// Serve creates a http.Server and calls ServeTLS with the given listener,
// certFile and keyFile. It wraps handler with RecoverAndLogHandler and a
// handler, which limits the max body size to config.MaxBodyBytes.
//
// NOTE: This function blocks - you may want to call it in a go-routine.
func ServeTLS(
	listener net.Listener,
	handler http.Handler,
	certFile, keyFile string,
	logger log.Logger,
	config *Config,
) error {
	logger.Info("serve tls", "msg", log.NewLazySprintf("Starting RPC HTTPS server on %s (cert: %q, key: %q)",
		listener.Addr(), certFile, keyFile))
	s := &http.Server{
		Handler:           RecoverAndLogHandler(maxBytesHandler{h: handler, n: config.MaxBodyBytes}, logger),
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
	}
	err := s.ServeTLS(listener, certFile, keyFile)

	logger.Error("RPC HTTPS server stopped", "err", err)
	return err
}

// WriteRPCResponseHTTPError marshals res as JSON (with indent) and writes it
// to w.
//
// source: https://www.jsonrpc.org/historical/json-rpc-over-http.html
func WriteRPCResponseHTTPError(
	w http.ResponseWriter,
	httpCode int,
	res types.RPCResponse,
) error {
	if res.Error == nil {
		panic("tried to write http error response without RPC error")
	}

	jsonBytes, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	_, err = w.Write(jsonBytes)
	return err
}

// WriteRPCResponseHTTP marshals res as JSON (with indent) and writes it to w.
func WriteRPCResponseHTTP(w http.ResponseWriter, res ...types.RPCResponse) error {
	return writeRPCResponseHTTP(w, []httpHeader{}, res...)
}

// WriteCacheableRPCResponseHTTP marshals res as JSON (with indent) and writes
// it to w. Adds cache-control to the response header and sets the expiry to
// one day.
func WriteCacheableRPCResponseHTTP(w http.ResponseWriter, res ...types.RPCResponse) error {
	return writeRPCResponseHTTP(w, []httpHeader{{"Cache-Control", "public, max-age=86400"}}, res...)
}

type httpHeader struct {
	name  string
	value string
}

func writeRPCResponseHTTP(w http.ResponseWriter, headers []httpHeader, res ...types.RPCResponse) error {
	var v interface{}
	if len(res) == 1 {
		v = res[0]
	} else {
		v = res
	}

	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	w.Header().Set("Content-Type", "application/json")
	for _, header := range headers {
		w.Header().Set(header.name, header.value)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	return err
}

// -----------------------------------------------------------------------------

// RecoverAndLogHandler wraps an HTTP handler, adding error logging.
// If the inner function panics, the outer function recovers, logs, sends an
// HTTP 500 error response.
func RecoverAndLogHandler(handler http.Handler, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the ResponseWriter to remember the status
		rww := &responseWriterWrapper{-1, w}
		begin := time.Now()

		rww.Header().Set("X-Server-Time", strconv.FormatInt(begin.Unix(), 10))

		defer handlePanicDuringRecovery(w)
		defer handlePanicInHandler(rww, r, logger, begin)

		handler.ServeHTTP(rww, r)
	})
}

func handlePanicDuringRecovery(w http.ResponseWriter) {
	if e := recover(); e != nil {
		fmt.Fprintf(os.Stderr, "Panic during RPC panic recovery: %v\n%v\n", e, string(debug.Stack()))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func handlePanicInHandler(rww *responseWriterWrapper, r *http.Request, logger log.Logger, begin time.Time) {
	if e := recover(); e != nil {
		handleRPCResponseOrError(e, rww, logger)
		logRequestDetails(rww, r, logger, begin)
	}
}

func handleRPCResponseOrError(e interface{}, rww *responseWriterWrapper, logger log.Logger) {
	if res, ok := e.(types.RPCResponse); ok {
		if wErr := WriteRPCResponseHTTP(rww, res); wErr != nil {
			logger.Error("failed to write response", "err", wErr)
		}
	} else {
		err := normalizeError(e)
		logger.Error("panic in RPC HTTP handler", "err", e, "stack", string(debug.Stack()))
		res := types.RPCInternalError(types.JSONRPCIntID(-1), err)
		if wErr := WriteRPCResponseHTTPError(rww, http.StatusInternalServerError, res); wErr != nil {
			logger.Error("failed to write response", "err", wErr)
		}
	}
}

func normalizeError(e interface{}) error {
	switch e := e.(type) {
	case error:
		return e
	case string:
		return errors.New(e)
	case fmt.Stringer:
		return errors.New(e.String())
	default:
		return errors.New("unknown panic")
	}
}

func logRequestDetails(rww *responseWriterWrapper, r *http.Request, logger log.Logger, begin time.Time) {
	durationMS := time.Since(begin).Nanoseconds() / 1000000
	if rww.Status == -1 {
		rww.Status = 200
	}
	logger.Debug("served RPC HTTP response",
		"method", r.Method,
		"url", r.URL,
		"status", rww.Status,
		"duration", durationMS,
		"remoteAddr", r.RemoteAddr,
	)
}

// Remember the status for logging.
type responseWriterWrapper struct {
	Status int
	http.ResponseWriter
}

func (w *responseWriterWrapper) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

// implements http.Hijacker.
func (w *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

type maxBytesHandler struct {
	h http.Handler
	n int64
}

func (h maxBytesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.n)
	h.h.ServeHTTP(w, r)
}

// Listen starts a new net.Listener on the given address.
// It returns an error if the address is invalid or the call to Listen() fails.
func Listen(addr string, maxOpenConnections int) (listener net.Listener, err error) {
	parts := strings.SplitN(addr, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf(
			"invalid listening address %s (use fully formed addresses, including the tcp:// or unix:// prefix)",
			addr,
		)
	}
	proto, addr := parts[0], parts[1]
	listener, err = net.Listen(proto, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %v: %v", addr, err)
	}
	if maxOpenConnections > 0 {
		listener = netutil.LimitListener(listener, maxOpenConnections)
	}

	return listener, nil
}
