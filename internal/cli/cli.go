package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"castspeak/internal/speak"
)

const usage = `Usage: castspeak <command> [flags]

Commands:
  serve          Start the HTTP server
  speak          Speak text on a Cast device
  devices        List Cast devices on the network
  devices saved  Show saved devices (no network)
  devices forget Delete saved devices file
  scan           Scan subnet for Cast devices (no mDNS)
  volume         Set device volume
  mute           Mute a device
  unmute         Unmute a device
  stop           Stop media on a device
  status         Show device status
  play           Play an audio URL on a device

Run 'castspeak <command> --help' for command-specific flags.`

func PrintUsage() {
	fmt.Fprintln(os.Stderr, usage)
}

// deviceFlags adds the common --device, --uuid, --host, and --timeout flags to a FlagSet.
func deviceFlags(fs *flag.FlagSet) (device, uuid, host *string, timeout *int) {
	device = fs.String("device", "", "Device name")
	uuid = fs.String("uuid", "", "Device UUID")
	host = fs.String("host", "", "Device address as host or host:port (skips mDNS discovery, default port 8009)")
	timeout = fs.Int("timeout", 5, "Discovery timeout in seconds")
	return
}

// requireDevice exits with an error if no device identifier is set.
func requireDevice(fs *flag.FlagSet, device, uuid, host string) {
	if device == "" && uuid == "" && host == "" {
		fmt.Fprintln(os.Stderr, "Error: --device, --uuid, or --host is required")
		fs.Usage()
		os.Exit(1)
	}
}

func timeoutCtx(seconds int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
}

func RunDevices(args []string) error {
	// Handle subcommands
	if len(args) > 0 {
		switch args[0] {
		case "saved":
			return runDevicesSaved()
		case "forget":
			return runDevicesForget()
		}
	}

	fs := flag.NewFlagSet("devices", flag.ExitOnError)
	timeout := fs.Int("timeout", 5, "Discovery timeout in seconds")
	fs.Parse(args)

	ctx, cancel := timeoutCtx(*timeout)
	defer cancel()

	devices, err := speak.ListDevices(ctx)
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		fmt.Println("No devices found.")
		return nil
	}

	printDevices(devices)
	return nil
}

func runDevicesSaved() error {
	devices, err := speak.LoadSavedDevices()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		fmt.Println("No saved devices.")
		return nil
	}
	printDevices(devices)
	return nil
}

func runDevicesForget() error {
	if err := speak.ForgetDevices(); err != nil {
		return err
	}
	fmt.Println("Saved devices removed.")
	return nil
}

func printDevices(devices []speak.Device) {
	for _, d := range devices {
		model := d.Model
		if model == "" {
			model = "unknown"
		}
		fmt.Printf("%-30s  %s:%d  %s  %s\n", d.Name, d.Addr, d.Port, d.UUID, model)
	}
}

func RunScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	timeout := fs.Int("timeout", 10, "Scan timeout in seconds")
	save := fs.Bool("save", false, "Save discovered devices to config file")
	fs.Parse(args)

	ctx, cancel := timeoutCtx(*timeout)
	defer cancel()

	fmt.Println("Scanning local network for Cast devices...")
	devices, err := speak.ScanDevices(ctx)
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		fmt.Println("No Cast devices found.")
		return nil
	}

	printDevices(devices)

	if *save {
		if err := speak.SaveDevices(devices); err != nil {
			return fmt.Errorf("save devices: %w", err)
		}
		path, err := speak.SavedDevicesPath()
		if err != nil {
			path = "(unknown path)"
		}
		fmt.Printf("Saved %d device(s) to %s\n", len(devices), path)
	}
	return nil
}

func RunSpeak(args []string) error {
	fs := flag.NewFlagSet("speak", flag.ExitOnError)
	device, uuid, host, timeout := deviceFlags(fs)
	text := fs.String("text", "", "Text to speak (required)")
	language := fs.String("language", "en", "Language code")
	fs.Parse(args)

	if *text == "" {
		fmt.Fprintln(os.Stderr, "Error: --text is required")
		fs.Usage()
		os.Exit(1)
	}
	requireDevice(fs, *device, *uuid, *host)

	ctx, cancel := timeoutCtx(*timeout)
	defer cancel()

	deviceName, chunks, err := speak.Speak(ctx, *text, *device, *uuid, *host, *language)
	if err != nil {
		return err
	}

	fmt.Printf("Spoke %d chunk(s) on %q\n", chunks, deviceName)
	return nil
}

func RunVolume(args []string) error {
	fs := flag.NewFlagSet("volume", flag.ExitOnError)
	device, uuid, host, timeout := deviceFlags(fs)
	level := fs.Float64("level", -1, "Volume level 0.0–1.0 (required)")
	fs.Parse(args)

	requireDevice(fs, *device, *uuid, *host)
	if *level < 0 || *level > 1 {
		fmt.Fprintln(os.Stderr, "Error: --level must be between 0.0 and 1.0")
		fs.Usage()
		os.Exit(1)
	}

	ctx, cancel := timeoutCtx(*timeout)
	defer cancel()

	if err := speak.SetVolume(ctx, *device, *uuid, *host, float32(*level)); err != nil {
		return err
	}
	fmt.Printf("Volume set to %.0f%%\n", *level*100)
	return nil
}

func RunMute(args []string, muted bool) error {
	action := "Muted"
	name := "mute"
	if !muted {
		action = "Unmuted"
		name = "unmute"
	}

	fs := flag.NewFlagSet(name, flag.ExitOnError)
	device, uuid, host, timeout := deviceFlags(fs)
	fs.Parse(args)

	requireDevice(fs, *device, *uuid, *host)

	ctx, cancel := timeoutCtx(*timeout)
	defer cancel()

	if err := speak.SetMuted(ctx, *device, *uuid, *host, muted); err != nil {
		return err
	}
	fmt.Printf("%s device\n", action)
	return nil
}

func RunStop(args []string) error {
	fs := flag.NewFlagSet("stop", flag.ExitOnError)
	device, uuid, host, timeout := deviceFlags(fs)
	fs.Parse(args)

	requireDevice(fs, *device, *uuid, *host)

	ctx, cancel := timeoutCtx(*timeout)
	defer cancel()

	if err := speak.Stop(ctx, *device, *uuid, *host); err != nil {
		return err
	}
	fmt.Println("Stopped")
	return nil
}

func RunStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	device, uuid, host, timeout := deviceFlags(fs)
	fs.Parse(args)

	requireDevice(fs, *device, *uuid, *host)

	ctx, cancel := timeoutCtx(*timeout)
	defer cancel()

	deviceName, status, err := speak.Status(ctx, *device, *uuid, *host)
	if err != nil {
		return err
	}

	fmt.Printf("Device:  %s\n", deviceName)
	fmt.Printf("App:     %s\n", status.App)
	fmt.Printf("Idle:    %t\n", status.Idle)
	fmt.Printf("Volume:  %.0f%%\n", status.Volume*100)
	fmt.Printf("Muted:   %t\n", status.Muted)
	if status.PlayerState != "" {
		fmt.Printf("State:   %s\n", status.PlayerState)
	}
	if status.MediaID != "" {
		fmt.Printf("Media:   %s\n", status.MediaID)
	}
	return nil
}

func RunPlay(args []string) error {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	device, uuid, host, timeout := deviceFlags(fs)
	url := fs.String("url", "", "Media URL to play (required)")
	fs.Parse(args)

	requireDevice(fs, *device, *uuid, *host)
	if *url == "" {
		fmt.Fprintln(os.Stderr, "Error: --url is required")
		fs.Usage()
		os.Exit(1)
	}

	ctx, cancel := timeoutCtx(*timeout)
	defer cancel()

	if err := speak.PlayURL(ctx, *device, *uuid, *host, *url); err != nil {
		return err
	}
	fmt.Println("Playing")
	return nil
}
