/*
 * (C) Copyright 2014 Nuxeo SA (http://nuxeo.com/) and contributors.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the GNU Lesser General Public License
 * (LGPL) version 2.1 which accompanies this distribution, and is available at
 * http://www.gnu.org/licenses/lgpl-2.1.html
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * Lesser General Public License for more details.
 *
 * Contributors:
 *     nuxeo.io Team
 */

package main

import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	"regexp"
	"strings"
	"time"
	"os/exec"
)

const LIMIT_TIME time.Duration = 12 * time.Hour

// A watcher loads and watch the etcd hierarchy for services.
type watcher struct {
	client   *etcd.Client
	config   *Config
	services map[string]*ServiceCluster
}

// Constructor for a new watcher
func NewEtcdWatcher(config *Config, services map[string]*ServiceCluster) (*watcher, error) {
	client, err := config.getEtcdClient()

	if err != nil {
		return nil, err
	}
	return &watcher{client, config, services}, nil
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
	glog.Error("passer par la")
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

func (w *watcher) getServiceForNode(node *etcd.Node) string {
	r := regexp.MustCompile(w.config.servicePrefix + "/(.*)(/.*)*")
	return strings.Split(r.FindStringSubmatch(node.Key)[1], "/")[0]
}

func (w *watcher) getServiceIndexForNode(node *etcd.Node) string {
	r := regexp.MustCompile(w.config.servicePrefix + "/(.*)(/.*)*")
	return strings.Split(r.FindStringSubmatch(node.Key)[1], "/")[1]
}

func (w *watcher) RemoveEnv(serviceName string) {
	delete(w.services, serviceName)
}

func (w *watcher) checkServiceAccess(node *etcd.Node, action string) {
	serviceName := w.getServiceForNode(node)

	// Get service's root node instead of changed node.
	serviceNode, _ := w.client.Get(w.config.servicePrefix+"/"+serviceName, true, true)

	for _, indexNode := range serviceNode.Node.Nodes {

		serviceIndex := w.getServiceIndexForNode(indexNode)
		serviceKey := w.config.servicePrefix + "/" + serviceName + "/" + serviceIndex
		lastAccessKey := serviceKey + "/lastAccess"

		response, err := w.client.Get(lastAccessKey, false, false)

		if err == nil {

			if w.services[serviceName] == nil {
				w.services[serviceName] = &ServiceCluster{}
			}

			service := &Service{}
			service.index = serviceIndex
			service.nodeKey = serviceKey

			if action == "delete" || action == "expire" {
				w.RemoveEnv(serviceName)
				return
			}

			lastAccess := response.Node.Value

			lastAccessTime, err := time.Parse(lastAccess, lastAccess)

			if err != nil {
				glog.Errorf("%s", err)
				break
			}

			if !time.Now().After(lastAccessTime.Add(LIMIT_TIME)) {
				actualService := w.services[serviceName].Get(service.index)
				if actualService != nil {
					_, err := exec.Command("fleetctl", "start", serviceName).Output()
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
