package main

import (
	"github.com/arkenio/goarken"
	"github.com/golang/glog"
	"os"
	"os/exec"
	"time"
)

type Passivator struct {
	Config  *Config
	Watcher *goarken.Watcher
	Stop    chan interface{}
}

func NewPassivator(w *goarken.Watcher, c *Config, stop chan interface{}) *Passivator {
	return &Passivator{Config: c, Watcher: w, Stop: stop}
}

func (p *Passivator) run() {
	cronDuration := p.Config.cronDuration
	interval := time.Duration(cronDuration) * time.Minute
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-p.Stop:
			return
		case <-ticker.C:

			for _, serviceCluster := range p.Watcher.Services {
				p.passivateServiceIfNeeded(serviceCluster)
			}

			ticker = time.NewTicker(interval)
		}
	}
}

func (p *Passivator) passivateServiceIfNeeded(serviceCluster *goarken.ServiceCluster) {

	service, err := serviceCluster.Next()
	if err != nil {
		//No active instance, no need to passivate
		return
	}

	// Checking if the service should be passivated or not
	if p.hasToBePassivated(service) {
		etcd, err := p.Config.getEtcdClient()

		statusKey := service.NodeKey + "/status"

		// TODO could expose an API in goarken package
		responseCurrent, error := etcd.Set(statusKey+"/current", goarken.PASSIVATED_STATUS, 0)
		if error != nil && responseCurrent == nil {
			glog.Errorf("Setting status current to 'passivated' has failed for Service "+service.Name+": %s", err)
		}

		response, error := etcd.Set(statusKey+"/expected", goarken.PASSIVATED_STATUS, 0)
		if error != nil && response == nil {
			glog.Errorf("Setting status expected to 'passivated' has failed for Service "+service.Name+": %s", err)
		}

		// TODO replace with Rest API
		cmd := exec.Command("/usr/bin/fleetctl", "--endpoint="+p.Config.etcdAddress, "stop", service.UnitName())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()

		if err != nil {
			glog.Errorf("Service "+service.Name+" passivation has failed: %s", err)
			return
		}
		glog.Infof("Service %s passivated", service.Name)

	}

}

func (p *Passivator) hasToBePassivated(service *goarken.Service) bool {

	parameter := p.Config.passiveLimitDuration
	passiveLimitDuration := time.Duration(parameter) * time.Hour

	return service.StartedSince() != nil &&
		time.Now().After(service.StartedSince().Add(passiveLimitDuration))
}
