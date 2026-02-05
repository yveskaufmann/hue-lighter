package events

import (
	"errors"
	"fmt"
	"net"
	"os"

	"com.github.yveskaufmann/hue-lighter/internal/services/light_automation"
	log "github.com/sirupsen/logrus"
)

type ExternalEventService struct {
	logger          *log.Entry
	lightAutomation *light_automation.Service
	listener        net.Listener
	stopChan        chan struct{}
}

func NewExternalEventService(lightAutomation *light_automation.Service, logger *log.Entry, stopChan chan struct{}) *ExternalEventService {
	return &ExternalEventService{
		logger:          logger.WithField("component", "ExternalEventService"),
		lightAutomation: lightAutomation,
		stopChan:        stopChan,
	}
}

func (s *ExternalEventService) Start() error {

	listener, err := net.Listen("unix", SOCKET_HUE_LIGHTER_EVENTS)
	if err != nil {
		return fmt.Errorf("failed to start Unix socket listener: %w", err)
	}
	s.listener = listener

	go func() {
		defer func() {
			s.logger.Info("Closing Unix socket listener")
			s.listener.Close()
			os.Remove(SOCKET_HUE_LIGHTER_EVENTS)
		}()

		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					s.logger.Info("Unix socket listener closed, stopping event loop")
					return
				}
				s.logger.WithError(err).Error("Failed to accept connection on Unix socket")
				continue

			}

			s.logger.Printf("Listening for events on Unix socket: %q", SOCKET_HUE_LIGHTER_EVENTS)

			buf := make([]byte, 128)
			defer conn.Close()
			n, _ := conn.Read(buf)
			if string(buf[:n]) == EVENT_TYPE_SHUTDOWN {
				s.logger.Info("Received shutdown event, stopping light automation service")
				err := s.lightAutomation.StopAndTurnOffLights()
				if err != nil {
					s.logger.WithError(err).Error("Failed to stop and turn off lights")
				}

				if s.stopChan != nil {
					s.stopChan <- struct{}{}
				}

				if err != nil {
					s.logger.WithError(err).Error("Failed to stop light automation service")
				}
				return
			}

		}
	}()

	s.logger.Info("Starting External Event Service")
	return nil
}

func (s *ExternalEventService) StopAndTurnOffLights() error {
	conn, err := net.Dial("unix", SOCKET_HUE_LIGHTER_EVENTS)
	if err != nil {
		return fmt.Errorf("failed to connect to Unix socket: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(EVENT_TYPE_SHUTDOWN))
	if err != nil {
		return fmt.Errorf("failed to send shutdown event: %w", err)
	}

	return nil
}

func (s *ExternalEventService) Stop() error {
	s.logger.Info("Stopping External Event Service")

	if s.listener != nil {
		s.logger.Info("Closing Unix socket listener")
		s.listener.Close()
	}

	return nil
}
