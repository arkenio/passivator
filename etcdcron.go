/*
 * (C) Copyright 2014 Nuxeo SA (http://nuxeo.com/) and contributors.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Apache License Version 2.0
 * which accompanies this distribution, and is available at
 * http://www.apache.org/licenses/
 *
 * See the Apache Licence for more details.
 *
 * Contributors:
 *     nuxeo.io Team
 */

package main

import (
	"github.com/coreos/go-etcd/etcd"
	"time"
	"github.com/golang/glog"
	"os/exec"
	"strconv"
	"os"
)

type EtcdCron struct {
	client   *etcd.Client
	config   *Config
}

func NewEtcdCron(config *Config) (*EtcdCron, error) {
	client, err := config.getEtcdClient()
	if err != nil {
		return nil, err
	}
	return &EtcdCron{client, config}, nil
}

func (etcdcron *EtcdCron) init() {
	etcdcron.start()
}

func (etcdcron *EtcdCron) start() {
	cronDuration, _ := strconv.Atoi(etcdcron.config.cronDuration)
	interval := time.Duration(cronDuration) * time.Second
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		// Check every 5 minutes all services lastAccess etcd date entry
		response, err := etcdcron.client.Get(etcdcron.config.servicePrefix, true, true)
		if err == nil {
			for _, serviceNode := range response.Node.Nodes {
				etcdcron.checkServiceAccess(serviceNode, response.Action)
			}
		}
		ticker = time.NewTicker(interval)
	}
}

func (etcdcron *EtcdCron) checkServiceAccess(node *etcd.Node, action string) {
	serviceName := etcdcron.config.getServiceForNode(node, etcdcron.config)

	// Get service's root node instead of changed node.
	serviceNode, _ := etcdcron.client.Get(etcdcron.config.servicePrefix+"/"+serviceName, true, true)

	for _, indexNode := range serviceNode.Node.Nodes {

		serviceIndex := etcdcron.config.getServiceIndexForNode(indexNode, etcdcron.config)
		serviceKey := etcdcron.config.servicePrefix + "/" + serviceName + "/" + serviceIndex
		lastAccessKey := serviceKey + "/lastAccess"
		statusKey := serviceKey + "/status"

		response, err := etcdcron.client.Get(serviceKey, true, true)

		if err == nil {

			service := &Service{}
			service.index = serviceIndex
			service.nodeKey = serviceKey
			service.name = "nxio."+serviceName+"."+serviceIndex+".service"

			for _, node := range response.Node.Nodes {
				switch node.Key {
				case statusKey:
					service.status = &Status{}
				for _, subNode := range node.Nodes {
					switch subNode.Key {
					case statusKey + "/alive":
						service.status.alive = subNode.Value
					case statusKey + "/current":
						service.status.current = subNode.Value
					case statusKey + "/expected":
						service.status.expected = subNode.Value
					}
				}
				case lastAccessKey:
					lastAccess := node.Value
					lastAccessTime, err := time.Parse(TIME_FORMAT, lastAccess)
					if err != nil {
						glog.Errorf("Error parsing last access date with service %s: %s", service.name, err)
						break
					}
					service.lastAccess = &lastAccessTime
				}
			}

			parameter, _ := strconv.Atoi(etcdcron.config.passiveLimitDuration)
			passiveLimitDuration := time.Duration(parameter) * time.Hour

			// Checking if the service should be passivated or not
			if service.lastAccess != nil && service.status != nil {
				if etcdcron.hasToBePassivated(service, passiveLimitDuration) {
					responseCurrent, error := etcdcron.client.Set(statusKey+"/current", PASSIVATED_STATUS, 0)
					if error != nil && responseCurrent == nil {
						glog.Errorf("Setting status current to 'passivated' has failed for Service "+service.name+": %s", err)
					}
					response, error := etcdcron.client.Set(statusKey+"/expected", PASSIVATED_STATUS, 0)
					if error != nil && response == nil {
						glog.Errorf("Setting status expected to 'passivated' has failed for Service "+service.name+": %s", err)
					}
					cmd := exec.Command("/usr/bin/fleetctl --endpoint=" + etcdcron.config.etcdAddress + " stop " + service.name)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Run()
					if err != nil {
						glog.Errorf("Service "+service.name+" passivation has failed: %s", err)
						break
					}
					glog.Infof("Service %s passivated", service.name)
				}
			}
		}
	}
}

func (etcdcron *EtcdCron) hasToBePassivated(service *Service, passiveLimitDuration time.Duration) bool {
	return time.Now().After(service.lastAccess.Add(passiveLimitDuration)) && service.status.current == STARTED_STATUS
}
