package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"castspeak/internal/speak"
)

func handleDevices(w http.ResponseWriter, r *http.Request) {
	timeout := speak.DefaultDiscoveryTimeout
	if t := r.URL.Query().Get("timeout"); t != "" {
		if secs, err := strconv.Atoi(t); err == nil && secs > 0 {
			timeout = time.Duration(secs) * time.Second
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	devices, err := speak.ListDevices(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	resp := DevicesResponse{Devices: make([]DeviceInfo, len(devices))}
	for i, d := range devices {
		resp.Devices[i] = DeviceInfo{
			Name:  d.Name,
			UUID:  d.UUID,
			Addr:  d.Addr,
			Port:  d.Port,
			Model: d.Model,
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleSpeak(w http.ResponseWriter, r *http.Request) {
	var req SpeakRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON body"})
		return
	}

	// Use the request context directly — the chi Timeout middleware (60s) is the ceiling.
	// Discovery + cast playback can take longer than the 5s discovery timeout.
	deviceName, chunks, err := speak.Speak(r.Context(), req.Text, req.DeviceName, req.DeviceUUID, req.Language)
	if err != nil {
		log.Printf("speak error: %v", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, SpeakResponse{
		Status: "ok",
		Device: deviceName,
		Chunks: chunks,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}
