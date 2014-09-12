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
	"os"
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
	serviceNode, err := w.client.Get(w.config.servicePrefix+"/"+serviceName, true, true)

	if (err == nil) {

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

				// Checking if the service should be re-activated or not
				if service.lastAccess != nil && service.status != nil {
					if w.hasToBeActivated(service) {
						cmd := exec.Command("/usr/bin/fleetctl", "--endpoint="+w.config.etcdAddress, "start", service.name)
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						err := cmd.Run()
						if err != nil {
							glog.Errorf("Service "+service.name+" restart has failed: %s", err)
							break
						}
						glog.Infof("Service %s restarted", service.name)
					}
				}
			}
		}
	} else {
		glog.Errorf("Unable to get information for service %s from etcd",serviceName)
	}
}

func (watcher *watcher) hasToBeActivated(service *Service) bool {
	return service.status.expected == STARTED_STATUS && service.status.current == STOPPED_STATUS
}

