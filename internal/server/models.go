package server

// DevicesResponse is the JSON response for GET /devices.
type DevicesResponse struct {
	Devices []DeviceInfo `json:"devices"`
}

// DeviceInfo represents a discovered Cast device.
type DeviceInfo struct {
	Name  string `json:"name"`
	UUID  string `json:"uuid"`
	Addr  string `json:"addr"`
	Port  int    `json:"port"`
	Model string `json:"model"`
}

// SpeakRequest is the JSON body for POST /speak.
type SpeakRequest struct {
	Text       string `json:"text"`
	DeviceName string `json:"device_name"`
	DeviceUUID string `json:"device_uuid"`
	Host       string `json:"host"`
	Language   string `json:"language"`
}

// SpeakResponse is the JSON response for POST /speak.
type SpeakResponse struct {
	Status string `json:"status"`
	Device string `json:"device"`
	Chunks int    `json:"chunks"`
}

// ErrorResponse is returned for error cases.
type ErrorResponse struct {
	Error string `json:"error"`
}
