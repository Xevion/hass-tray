package internal

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/getlantern/systray"
	dotenv "github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	ga "saml.dev/gome-assistant"
)

var ()

type TrayApp struct {
	doorIdentifier string
	log            *slog.Logger
	stateChannel   chan string
	app            *ga.App
	service        *ga.Service
}

// Status will return the operational status of the service
func (ta *TrayApp) Status() Status {
	return StatusUnknown
}

func (ta *TrayApp) State() string {
	// TODO: Implement this method
	return ""
}

func (ta *TrayApp) Connected() bool {
	// TODO: Implement this method
	return false
}

func (ta *TrayApp) Reload() error {
	// TODO: Implement this method
	return nil
}

func (ta *TrayApp) Pause() error {
	// TODO: Implement this method
	return nil
}

func (ta *TrayApp) Resume() error {
	// TODO: Implement this method
	return nil
}

func NewApp() *TrayApp {
	// Connect to Home Assistant
	app, err := ga.NewApp(ga.NewAppRequest{
		IpAddress:        "home.imfucked.lol", // Replace with your Home Assistant IP Address
		HAAuthToken:      os.Getenv("HA_AUTH_TOKEN"),
		HomeZoneEntityId: "zone.home",
		Port:             "443",
		Secure:           true,
	})
	if err != nil {
		slog.Error("Error connecting to Home Assistant", "error", err)
		os.Exit(1)
	}

	service := app.GetService()

	return &TrayApp{
		app:            app,
		service:        service,
		stateChannel:   make(chan string),
		doorIdentifier: "binary_sensor.bedroom_door_opening",
	}
}

var (
	//go:embed "resources/*.ico"
	icons embed.FS
)

func (ta *TrayApp) HandleState(newState string) {
	switch newState {
	case "on":
		ta.stateChannel <- "open"
	case "off":
		ta.stateChannel <- "closed"
	default:
		slog.Error("unknown state encountered", "newState", newState)
		ta.stateChannel <- "unknown"
	}
}

func (ta *TrayApp) setupHomeAssistant() {
	var err error

	// Get the initial state
	state, err := ta.app.GetState().Get(ta.doorIdentifier)
	if err != nil {
		slog.Error("Unable to get initial state", "error", err)
	} else {
		slog.Debug("Initial State Received")
		ta.HandleState(state.State)
	}

	ta.app.RegisterEntityListeners(ga.
		NewEntityListener().
		EntityIds(ta.doorIdentifier).
		Call(func(service *ga.Service, state ga.State, sensor ga.EntityData) {
			slog.Debug("Event Received", "identifier", ta.doorIdentifier, "sensor", sensor)
			ta.HandleState(sensor.ToState)
		}).
		Build())

	ta.app.Start()

	slog.Warn("Home Assistant thread died")
	ta.stateChannel <- "unknown"
}

func (ta *TrayApp) Start() {
	dotenv.Load()

	slog.SetDefault(slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	)))
	// binfo, err := buildinfo.
	slog.Info("Startup", "runtime", runtime.Version(), "os", runtime.GOOS, "arch", runtime.GOARCH, "pid", os.Getpid())

	go ta.setupHomeAssistant()
	systray.Run(ta.onReady, func() {})
}

func (ta *TrayApp) onReady() {
	systray.SetTitle("door-tray")
	systray.SetTooltip("Setting up...")
	menuQuit := systray.AddMenuItem("Quit", "Stops the application")
	menuOpenLogs := systray.AddMenuItem("Open Logs", "Opens the logs in the default editor")
	menuOpenLogs.Disable()

	// Load icons
	systray.SetIcon(getIcon("unknown"))

	// Handle Ctrl+C interrupt
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	signal.Notify(interruptChannel, os.Kill)

loop:
	for {
		select {
		case signal := <-interruptChannel:
			slog.Info("Received interrupt signal, quitting", "signal", signal)
			break loop
		case <-menuQuit.ClickedCh:
			slog.Info("Quit clicked")
			break loop
		case <-menuOpenLogs.ClickedCh:
			slog.Info("Open Logs clicked")
		case newState := <-ta.stateChannel:
			timeString := time.Now().Format("3:04 PM")
			if newState != "unknown" {
				systray.SetTooltip(fmt.Sprintf("%s as of %s", cases.Title(language.AmericanEnglish, cases.NoLower).String(newState), timeString))
				switch newState {
				case "open":
					systray.SetIcon(getIcon("open_fault"))
				case "closed":
					systray.SetIcon(getIcon("closed"))
				}
			} else {
				slog.Warn("Unknown state", "state", newState)
				systray.SetTooltip(fmt.Sprintf("Unknown as of %s", timeString))
				systray.SetIcon(getIcon("unknown"))
			}
		}
	}

	slog.Info("Cleaning up")
	systray.Quit()
	ta.app.Cleanup()
}
