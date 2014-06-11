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
	"github.com/golang/glog"
	"time"
	"os/exec"
	"strconv"
)

// A watcher loads and watch the etcd hierarchy for services.
type watcher struct {
	client   *etcd.Client
	config   *Config
}

// Constructor for a new watcher
func NewEtcdWatcher(config *Config) (*watcher, error) {
	client, err := config.getEtcdClient()

	if err != nil {
		return nil, err
	}
	return &watcher{client, config}, nil
}

//Init services watcher
func (w *watcher) init() {
	w.loadAndWatch(w.config.servicePrefix, w.checkServiceAccess)
}

// Loads and watch an etcd directory to register objects like services
// The register function is passed the etcd Node that has been loaded.
func (w *watcher) loadAndWatch(etcdDir string, registerFunc func(*etcd.Node, string)) {
	w.loadPrefix(etcdDir, registerFunc)
	updateChannel := make(chan *etcd.Response, 10)
	go w.watch(updateChannel, registerFunc)
	w.client.Watch(etcdDir, (uint64)(0), true, updateChannel, nil)
}

func (w *watcher) loadPrefix(etcDir string, registerFunc func(*etcd.Node, string)) {
	response, err := w.client.Get(etcDir, true, true)
	if err == nil {
		for _, serviceNode := range response.Node.Nodes {
			registerFunc(serviceNode, response.Action)
		}
	}
}

func (w *watcher) watch(updateChannel chan *etcd.Response, registerFunc func(*etcd.Node, string)) {
	for {
		response := <-updateChannel
		if response != nil {
			registerFunc(response.Node, response.Action)
		}
	}
}

func (w *watcher) checkServiceAccess(node *etcd.Node, action string) {
	serviceName := w.config.getServiceForNode(node, w.config)

	// Get service's root node instead of changed node.
	serviceNode, _ := w.client.Get(w.config.servicePrefix+"/"+serviceName, true, true)

	for _, indexNode := range serviceNode.Node.Nodes {

		serviceIndex := w.config.getServiceIndexForNode(indexNode, w.config)
		serviceKey := w.config.servicePrefix + "/" + serviceName + "/" + serviceIndex
		lastAccessKey := serviceKey + "/lastAccess"
		statusKey := serviceKey + "/status"

		response, err := w.client.Get(serviceKey, true, true)

		if err == nil {

			service := &Service{}
			service.index = serviceIndex
			service.nodeKey = serviceKey

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
						glog.Errorf("Error parsing last access date with service %s: %s", serviceName, err)
						break
					}
					service.lastAccess = &lastAccessTime
				}
			}

			parameter, _ := strconv.Atoi(w.config.passiveLimitDuration)
			passiveLimitDuration := time.Duration(parameter) * time.Hour


			// Checking if the service should be re-activated or not
			if service.lastAccess != nil && service.status != nil {
				if w.hasToBeActivated(service, passiveLimitDuration) {
					response, error := w.client.Set(statusKey+"/expected", STARTED_STATUS, 0)
					if error != nil && response == nil {
						glog.Errorf("Setting expected status to 'started' has failed for Service "+serviceName+": %s", err)
					}
					_, err := exec.Command("/bin/bash -c fleetctl start " + serviceName).Output()
					if err != nil {
						glog.Errorf("Service "+serviceName+" restart has failed: %s", err)
						break
					}
					glog.Infof("Service %s restarted", serviceName)
				}
			}
		}
	}
}

func (watcher *watcher) hasToBeActivated(service *Service, passiveLimitDuration time.Duration) bool {
	return !time.Now().After(service.lastAccess.Add(passiveLimitDuration)) && service.status.expected == PASSIVATED_STATUS
}

