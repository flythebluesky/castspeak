package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRouter_Routes(t *testing.T) {
	handler := New()

	tests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{"GET", "/devices", "", http.StatusOK},               // may return empty devices
		{"POST", "/speak", `{"text":"hi"}`, http.StatusOK},   // will fail at discovery but we test routing
		{"GET", "/nonexistent", "", http.StatusNotFound},      // chi returns 405 or 404
		{"DELETE", "/devices", "", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// We can't assert exact status for /devices and /speak since they depend
			// on network discovery. Just verify the routes exist (not 404/405).
			switch tt.path {
			case "/devices":
				if tt.method == "GET" && w.Code == http.StatusNotFound {
					t.Errorf("GET /devices returned 404, route not registered")
				}
			case "/speak":
				if tt.method == "POST" && w.Code == http.StatusNotFound {
					t.Errorf("POST /speak returned 404, route not registered")
				}
			case "/nonexistent":
				if w.Code != http.StatusNotFound {
					t.Errorf("got status %d, want 404", w.Code)
				}
			}
		})
	}
}

func TestHandleSpeak_InvalidJSON(t *testing.T) {
	handler := New()

	req := httptest.NewRequest("POST", "/speak", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want 400", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error == "" {
		t.Error("expected error message in response")
	}
}

func TestHandleSpeak_MissingText(t *testing.T) {
	handler := New()

	req := httptest.NewRequest("POST", "/speak",
		strings.NewReader(`{"device_name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// The speak package validates text is required and returns an error.
	// The handler should return a non-200 status.
	if w.Code == http.StatusOK {
		t.Error("expected error status for missing text")
	}
}

func TestHandleSpeak_MissingDevice(t *testing.T) {
	handler := New()

	req := httptest.NewRequest("POST", "/speak",
		strings.NewReader(`{"text":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Missing device should cause an error from the speak package.
	if w.Code == http.StatusOK {
		t.Error("expected error status for missing device")
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusCreated, map[string]string{"key": "value"})

	if w.Code != http.StatusCreated {
		t.Errorf("got status %d, want 201", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("body[key] = %q, want %q", body["key"], "value")
	}
}
