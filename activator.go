package main

import (
	"github.com/arkenio/goarken"
	"github.com/golang/glog"
	"os"
	"os/exec"
)

type Activator struct {
	Config  *Config
	Watcher *goarken.Watcher
	Stop    chan interface{}
}

func NewActivator(w *goarken.Watcher, c *Config, stop chan interface{}) *Activator {
	return &Activator{Config: c, Watcher: w, Stop: stop}
}

func (a *Activator) run() {

	updateChannel := a.Watcher.Listen()

	for {
		select {
		case <-a.Stop:
			return
		case serviceOrDomain := <-updateChannel:
			if service, ok := serviceOrDomain.(goarken.Service); ok {
				a.restartIfNeeded(&service)
			}
		}
	}

}

func (a *Activator) restartIfNeeded(service *goarken.Service) {

	if a.hasToBeRestarted(service) {
		//TODO Use fleet's REST API
		cmd := exec.Command("/usr/bin/fleetctl", "--endpoint="+a.Config.etcdAddress, "start", service.UnitName())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			glog.Errorf("Service "+service.Name+" restart has failed: %s", err)
			return
		}

		glog.Infof("Service %s restarted", service.Name)
	}
}

func (a *Activator) hasToBeRestarted(service *goarken.Service) bool {
	return service.LastAccess != nil &&
		service.Status != nil &&
		service.Status.Expected == goarken.STARTED_STATUS &&
		service.Status.Current == goarken.STOPPED_STATUS

}
