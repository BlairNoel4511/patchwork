package server

import (
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/patrickward/patchwork/config"
)

// LatencyMiddleware injects artificial latency based on a LatencyProfile.
// It supports fixed, uniform, and normal distributions.
func LatencyMiddleware(profile *config.LatencyProfile) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if profile != nil {
				d := sampleLatency(profile)
				if d > 0 {
					time.Sleep(d)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// sampleLatency returns a duration sampled from the given profile.
func sampleLatency(p *config.LatencyProfile) time.Duration {
	switch p.Distribution {
	case "uniform":
		span := p.MaxMs - p.MinMs
		if span <= 0 {
			return time.Duration(p.MinMs) * time.Millisecond
		}
		ms := p.MinMs + rand.Intn(span)
		return time.Duration(ms) * time.Millisecond
	case "normal":
		ms := rand.NormFloat64()*p.StdDevMs + p.MeanMs
		ms = math.Max(0, ms)
		return time.Duration(ms) * time.Millisecond
	default: // "fixed" or unset
		return time.Duration(p.FixedMs) * time.Millisecond
	}
}
