package main

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/getlantern/systray"
	dotenv "github.com/joho/godotenv"
	ga "saml.dev/gome-assistant"
)

var (
	doorIdentifier = "binary_sensor.bedroom_door_opening"
	service        *ga.Service
	log            *slog.Logger
	stateChannel   chan string
)

func HandleState(newState string) {
	switch newState {
	case "on":
		stateChannel <- "open"
	case "off":
		stateChannel <- "closed"
	default:
		slog.Error("unknown state encountered", "newState", newState)
		stateChannel <- "unknown"
	}
}

func setupHomeAssistant() {
	// Connect to Home Assistant
	app, err := ga.NewApp(ga.NewAppRequest{
		IpAddress:        "home.imfucked.lol", // Replace with your Home Assistant IP Address
		HAAuthToken:      os.Getenv("HA_AUTH_TOKEN"),
		HomeZoneEntityId: "zone.home",
		Port:             "443",
		Secure:           true,
	})
	if err != nil {
		log.Error("Error connecting to Home Assistant", "error", err)
		os.Exit(1)
	}
	defer func() {
		app.Cleanup()
		log.Debug("Deferred!")
	}()

	service = app.GetService()

	// Get the initial state
	state, err := app.GetState().Get(doorIdentifier)
	if err != nil {
		slog.Error("Unable to get initial state", "error", err)
	} else {
		slog.Debug("Initial State Received")
		HandleState(state.State)
	}

	app.RegisterEntityListeners(ga.
		NewEntityListener().
		EntityIds(doorIdentifier).
		Call(func(service *ga.Service, state ga.State, sensor ga.EntityData) {
			slog.Debug("Event Received", "identifier", doorIdentifier, "sensor", sensor)
			HandleState(sensor.ToState)
		}).
		Build())

	app.Start()
}

func main() {
	dotenv.Load()
	stateChannel = make(chan string)

	slog.SetDefault(slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	)))
	slog.Info("Starting hass-tray")

	go setupHomeAssistant()
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("hass-tray")
	systray.SetTooltip("Refreshed")
	menuQuit := systray.AddMenuItem("Quit", "Stops the application")
	menuOpenLogs := systray.AddMenuItem("Open Logs", "Opens the logs in the default editor")
	menuOpenLogs.Disable()

	// Load icons
	openIcon, err := os.ReadFile("open.ico")
	if err != nil {
		slog.Error("Unable to load icon", "error", err)
		os.Exit(1)
	}
	closedIcon, err := os.ReadFile("closed.ico")
	if err != nil {
		slog.Error("Unable to load icon", "error", err)
		os.Exit(1)
	}
	unknownIcon, err := os.ReadFile("unknown.ico")
	if err != nil {
		slog.Error("Unable to load icon", "error", err)
		os.Exit(1)
	} else {
		slog.Debug("Icons loaded")
	}
	systray.SetIcon(unknownIcon)

	// Handle Ctrl+C interrupt
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	signal.Notify(interruptChannel, os.Kill)

	for {
		select {
		case <-interruptChannel:
			slog.Info("Received interrupt signal, quitting")
			systray.Quit()
		case <-menuQuit.ClickedCh:
			slog.Info("Requesting exit")
			systray.Quit()
		case newState := <-stateChannel:
			if newState == "open" {
				systray.SetIcon(openIcon)
			} else if newState == "closed" {
				systray.SetIcon(closedIcon)
			} else {
				slog.Warn("Unknown state", "state", newState)
				systray.SetIcon(unknownIcon)
			}
		}
	}
}

func onExit() {

}
