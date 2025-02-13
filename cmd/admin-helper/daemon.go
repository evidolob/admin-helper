package main

import (
	"log"
	"net/http"
	"time"

	"github.com/crc-org/admin-helper/pkg/api"
	"github.com/crc-org/admin-helper/pkg/hosts"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

var InstallDaemon = &cobra.Command{
	Use:   "install-daemon",
	Short: "Install the daemon",
	RunE: func(_ *cobra.Command, _ []string) error {
		svc, err := svc()
		if err != nil {
			return err
		}
		if err := svc.Install(); err != nil {
			return err
		}
		return svc.Start()
	},
}

var UninstallDaemon = &cobra.Command{
	Use:   "uninstall-daemon",
	Short: "Uninstall the daemon",
	RunE: func(_ *cobra.Command, _ []string) error {
		svc, err := svc()
		if err != nil {
			return err
		}
		if err := svc.Stop(); err != nil {
			log.Println(err)
		}
		return svc.Uninstall()
	},
}

var Daemon = &cobra.Command{
	Use:   "daemon",
	Short: "Run as a daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon()
	},
}

func daemon() error {
	svc, err := svc()
	if err != nil {
		return err
	}
	return svc.Run()
}

func svc() (service.Service, error) {
	/*
	 * The following struct fields are used to set the service's Name,
	 * Display name, description and Arguments only when service  gets
	 * installed via the `admin-helper install-daemon` command for CRC
	 * in production these values are not used as the MSI installs the
	 * service
	 */
	svcConfig := &service.Config{
		Name:        "crcAdminHelper",
		DisplayName: "CRC Admin Helper",
		Description: "Perform administrative tasks for the user",
		Arguments:   []string{"daemon"},
	}
	prg := &program{}
	return service.New(prg, svcConfig)
}

type program struct{}

func (p *program) Start(s service.Service) error {
	go func() {
		logger, err := s.Logger(nil)
		if err != nil {
			log.Fatal(err)
		}
		ln, err := listen()
		if err != nil {
			_ = logger.Error(err)
			return
		}
		hosts, err := hosts.New()
		if err != nil {
			_ = logger.Error(err)
			return
		}
		s := &http.Server{
			Handler:      api.Mux(hosts),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		if err := s.Serve(ln); err != nil {
			_ = logger.Error(err)
			return
		}
	}()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}
