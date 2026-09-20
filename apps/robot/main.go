package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.bug.st/serial"
)

const (
	roombaBaudRate = 115200
	ledInterval    = 1 * time.Second
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s SERIAL_PORT", os.Args[0])
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, portPath string) error {
	port, err := serial.Open(portPath, &serial.Mode{BaudRate: roombaBaudRate})
	if err != nil {
		return fmt.Errorf("open serial port %q: %w", portPath, err)
	}
	defer port.Close()

	if _, err := (StartCommand{}).WriteTo(port); err != nil {
		return fmt.Errorf("send start command: %w", err)
	}
	if _, err := (SafeModeCommand{}).WriteTo(port); err != nil {
		return fmt.Errorf("send safe mode command: %w", err)
	}

	time.Sleep(1 * time.Second)

	ledsOn := false
	if err := writeLEDs(port, ledsOn); err != nil {
		return err
	}

	if _, err := (DriveCommand{
		VelocityMMPerSec: 0,
		RadiusMM:         DriveTurnClockwise,
	}).WriteTo(port); err != nil {
		return fmt.Errorf("drive: %w", err)
	}

	ticker := time.NewTicker(ledInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return shutdown(port)
		case <-ticker.C:
			ledsOn = !ledsOn
			if err := writeLEDs(port, ledsOn); err != nil {
				return err
			}
		}
	}
}

// shutdown runs serial commands synchronously, so cancellation is handled only
// after any in-progress command has completed.
func shutdown(port serial.Port) error {
	_, stopErr := (StopCommand{}).WriteTo(port)
	return stopErr
}

func writeLEDs(port serial.Port, on bool) error {
	command := LEDsCommand{}
	if on {
		command = LEDsCommand{
			LEDs:           LEDDebris | LEDSpot | LEDDock | LEDCheckRobot,
			PowerColor:     0,
			PowerIntensity: 255,
		}
	}

	if _, err := command.WriteTo(port); err != nil {
		return fmt.Errorf("set LEDs: %w", err)
	}
	return nil
}
