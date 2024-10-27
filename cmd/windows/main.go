package main

import (
	"log"
	"log/slog"
	"time"

	"internal"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

type WrapperService struct {
	service internal.Service
}

func (wrapper *WrapperService) Execute(args []string, requestChannel <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	const acceptedCommands = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	tick := time.Tick(5 * time.Second)

	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: acceptedCommands}

loop:
	for {
		select {
		case <-tick:
			slog.Debug("Tick Handled...!")
		case changeRequest := <-requestChannel:
			switch changeRequest.Cmd {
			case svc.Interrogate:
				slog.Debug("Interrogate Requested", "changeRequest", changeRequest)
				status <- changeRequest.CurrentStatus
			case svc.Stop, svc.Shutdown:
				slog.Warn("Shutdown Requested", "changeRequest", changeRequest)
				break loop
			case svc.Pause:
				wrapper.service.Pause()
				slog.Warn("Pause Requested", "changeRequest", changeRequest)
				status <- svc.Status{State: svc.Paused, Accepts: acceptedCommands}
			case svc.Continue:
				slog.Info("Continue Requested", "changeRequest", changeRequest)
				status <- svc.Status{State: svc.Running, Accepts: acceptedCommands}
			default:
				slog.Warn("Unexpected Change Request", "changeRequest", changeRequest)
			}
		}
	}

	slog.Info("Service Stopping")
	status <- svc.Status{State: svc.StopPending}
	return false, 1
}

func runService(name string, isDebug bool) {
	service := WrapperService{
		service: internal.NewApp(),
	}

	if isDebug {
		err := debug.Run(name, &service)
		if err != nil {
			log.Fatalln("Error running service in debug mode.")
		}
	} else {
		err := svc.Run(name, &service)
		if err != nil {
			log.Fatalln("Error running service in Service Control mode.")
		}
	}
}

func main() {

	runService("DoorTray", true)
}
