package cast

import (
	"fmt"

	"github.com/vishen/go-chromecast/application"
)

// withApp connects to a Cast device, runs fn, then closes the connection.
// stopOnClose controls whether to stop the active Cast app when closing.
func withApp(addr string, port int, stopOnClose bool, fn func(*application.Application) error) error {
	app := application.NewApplication(
		application.WithCacheDisabled(true),
	)
	if err := app.Start(addr, port); err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer app.Close(stopOnClose)
	return fn(app)
}

// PlayURLs connects to a Cast device and plays each URL sequentially.
func PlayURLs(addr string, port int, urls []string) error {
	return withApp(addr, port, true, func(app *application.Application) error {
		for i, u := range urls {
			if err := app.Load(u, 0, "audio/mpeg", false, false, false); err != nil {
				return fmt.Errorf("failed to play chunk %d: %w", i+1, err)
			}
		}
		return nil
	})
}

// SetVolume connects to a device and sets its volume level (0.0–1.0).
func SetVolume(addr string, port int, level float32) error {
	return withApp(addr, port, false, func(app *application.Application) error {
		return app.SetVolume(level)
	})
}

// SetMuted connects to a device and sets its mute state.
func SetMuted(addr string, port int, muted bool) error {
	return withApp(addr, port, false, func(app *application.Application) error {
		return app.SetMuted(muted)
	})
}

// Stop connects to a device and stops any active media.
func Stop(addr string, port int) error {
	return withApp(addr, port, false, func(app *application.Application) error {
		return app.Stop()
	})
}

// DeviceStatus holds the current state of a Cast device.
type DeviceStatus struct {
	App         string
	Idle        bool
	Volume      float32
	Muted       bool
	PlayerState string
	MediaID     string
}

// Status connects to a device and returns its current status.
func Status(addr string, port int) (DeviceStatus, error) {
	var ds DeviceStatus
	err := withApp(addr, port, false, func(app *application.Application) error {
		castApp, media, volume := app.Status()
		if castApp != nil {
			ds.App = castApp.DisplayName
			ds.Idle = castApp.IsIdleScreen
		}
		if volume != nil {
			ds.Volume = volume.Level
			ds.Muted = volume.Muted
		}
		if media != nil {
			ds.PlayerState = media.PlayerState
			ds.MediaID = media.Media.ContentId
		}
		return nil
	})
	return ds, err
}
