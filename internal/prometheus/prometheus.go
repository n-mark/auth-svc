package prometheus

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// latency
	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// rps / total requests
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// 5xx errors
	httpErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total HTTP 5xx responses",
		},
		[]string{"method", "path"},
	)
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         200,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		method := r.Method
		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(rw.status)

		// latency
		httpDuration.WithLabelValues(
			method,
			path,
			status,
		).Observe(duration)

		// total requests
		httpRequestsTotal.WithLabelValues(
			method,
			path,
			status,
		).Inc()

		// 5xx errors
		if rw.status >= 500 {
			httpErrorsTotal.WithLabelValues(
				method,
				path,
			).Inc()
		}
	})
}

var userIDRe = regexp.MustCompile(`^/user/\d+$`)

func normalizePath(path string) string {
	if strings.HasPrefix(path, "/user/") && userIDRe.MatchString(path) {
		return "/user/{id}"
	}
	return path
}
