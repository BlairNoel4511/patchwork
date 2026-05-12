package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/user/patchwork/config"
)

// RequestValidationMiddleware validates incoming requests against the route's validation config.
func RequestValidationMiddleware(cfg *config.RequestValidationConfig, next http.Handler) http.Handler {
	if cfg == nil || !cfg.IsEnabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content-Type check
		if len(cfg.RequireContentType) > 0 {
			ct := r.Header.Get("Content-Type")
			matched := false
			for _, allowed := range cfg.RequireContentType {
				if strings.Contains(ct, allowed) {
					matched = true
					break
				}
			}
			if !matched {
				rejectValidation(w, cfg)
				return
			}
		}

		// Body size + JSON field checks
		if cfg.MaxBodyBytes > 0 || len(cfg.RequireFields) > 0 || cfg.RejectUnknownFields {
			body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyLimit(cfg)))
			if err != nil {
				rejectValidation(w, cfg)
				return
			}
			if cfg.MaxBodyBytes > 0 && int64(len(body)) > cfg.MaxBodyBytes {
				rejectValidation(w, cfg)
				return
			}
			// Restore body for downstream handlers
			r.Body = io.NopCloser(bytes.NewReader(body))

			if len(cfg.RequireFields) > 0 || cfg.RejectUnknownFields {
				var obj map[string]json.RawMessage
				if err := json.Unmarshal(body, &obj); err != nil {
					rejectValidation(w, cfg)
					return
				}
				for _, f := range cfg.RequireFields {
					if _, ok := obj[f]; !ok {
						rejectValidation(w, cfg)
						return
					}
				}
				if cfg.RejectUnknownFields && len(cfg.AllowedFields) > 0 {
					allowed := make(map[string]struct{}, len(cfg.AllowedFields))
					for _, f := range cfg.AllowedFields {
						allowed[f] = struct{}{}
					}
					for k := range obj {
						if _, ok := allowed[k]; !ok {
							rejectValidation(w, cfg)
							return
						}
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func rejectValidation(w http.ResponseWriter, cfg *config.RequestValidationConfig) {
	w.WriteHeader(cfg.ResolvedStatusCode())
	_, _ = w.Write([]byte(cfg.ResolvedBody()))
}

func maxBodyLimit(cfg *config.RequestValidationConfig) int64 {
	if cfg.MaxBodyBytes > 0 {
		// Read one extra byte to detect overflow
		return cfg.MaxBodyBytes + 1
	}
	return 1 << 20 // 1 MB default cap for field inspection
}
