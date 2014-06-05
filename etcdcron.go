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
)

const INTERVAL_PERIOD time.Duration = 5 * time.Minute

type EtcdCron struct {
	client   *etcd.Client
	config   *Config
	services map[string]*ServiceCluster
}

func NewEtcdCron(config *Config, services map[string]*ServiceCluster) (*EtcdCron, error) {
	client, err := config.getEtcdClient()
	if err != nil {
		return nil, err
	}
	return &EtcdCron{client, config, services}, nil
}

func (etcdcron *EtcdCron) init() {
	etcdcron.start()
}

func (etcdcron *EtcdCron) start() {
	ticker := time.NewTicker(INTERVAL_PERIOD)
	for {
		<-ticker.C
		// Check every 5 minutes all services lastAccess etcd date entry
		response, err := etcdcron.client.Get(etcdcron.config.servicePrefix, true, true)
		if err == nil {
			for _, serviceNode := range response.Node.Nodes {
				etcdcron.checkServiceAccess(serviceNode, response.Action)
			}
		}
		ticker = time.NewTicker(INTERVAL_PERIOD)
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

			if etcdcron.services[serviceName] == nil {
				etcdcron.services[serviceName] = &ServiceCluster{}
			}

			service := &Service{}
			service.index = serviceIndex
			service.nodeKey = serviceKey

			if action == "delete" || action == "expire" {
				etcdcron.config.RemoveEnv(serviceName, etcdcron.services)
				return
			}

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
					lastAccess := response.Node.Value
					lastAccessTime, err := time.Parse(lastAccess, lastAccess)
					if err != nil {
						glog.Errorf("Error parsing last access date with service %s: %s", serviceName, err)
						break
					}
					service.lastAccess = lastAccessTime
				}
			}

			if time.Now().After(service.lastAccess.Add(LIMIT_TIME)) && service.status.current == STARTED_STATUS {
				actualService := etcdcron.services[serviceName].Get(service.index)
				if actualService != nil {
					_, error := etcdcron.client.Set(statusKey+"/current", PASSIVATED_STATUS, 0)
					if error == nil {
						glog.Errorf("Setting status current to passivated has failed for Service "+serviceName+": %s", err)
					}
					response, error := etcdcron.client.Set(statusKey+"/expected", PASSIVATED_STATUS, 0)
					if error == nil && response == nil {
						glog.Errorf("Setting status expected to passivated has failed for Service "+serviceName+": %s", err)
					}
					_, err := exec.Command("fleetctl", "stop", serviceName).Output()
					if err != nil {
						glog.Errorf("Service "+serviceName+" passivation has failed: %s", err)
						break
					}
					glog.Infof("Service %s passivated", serviceName)
				}
			}
		}
	}
}
