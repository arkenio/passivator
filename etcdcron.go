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
	"time"
	"github.com/golang/glog"
	"os/exec"
	"regexp"
	"strings"
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

func (etcdcron *EtcdCron) getServiceForNode(node *etcd.Node) string {
	r := regexp.MustCompile(etcdcron.config.servicePrefix + "/(.*)(/.*)*")
	return strings.Split(r.FindStringSubmatch(node.Key)[1], "/")[0]
}

func (etcdcron *EtcdCron) getServiceIndexForNode(node *etcd.Node) string {
	r := regexp.MustCompile(etcdcron.config.servicePrefix + "/(.*)(/.*)*")
	return strings.Split(r.FindStringSubmatch(node.Key)[1], "/")[1]
}

func (etcdcron *EtcdCron) RemoveEnv(serviceName string) {
	delete(etcdcron.services, serviceName)
}

func (etcdcron *EtcdCron) checkServiceAccess(node *etcd.Node, action string) {
	serviceName := etcdcron.getServiceForNode(node)

	// Get service's root node instead of changed node.
	serviceNode, _ := etcdcron.client.Get(etcdcron.config.servicePrefix+"/"+serviceName, true, true)

	for _, indexNode := range serviceNode.Node.Nodes {

		serviceIndex := etcdcron.getServiceIndexForNode(indexNode)
		serviceKey := etcdcron.config.servicePrefix + "/" + serviceName + "/" + serviceIndex
		lastAccessKey := serviceKey + "/lastAccess"

		response, err := etcdcron.client.Get(lastAccessKey, false, false)

		if err == nil {

			if etcdcron.services[serviceName] == nil {
				etcdcron.services[serviceName] = &ServiceCluster{}
			}

			service := &Service{}
			service.index = serviceIndex
			service.nodeKey = serviceKey

			if action == "delete" || action == "expire" {
				etcdcron.RemoveEnv(serviceName)
				return
			}

			lastAccess := response.Node.Value

			lastAccessTime, err := time.Parse(lastAccess, lastAccess)

			if err != nil {
				glog.Errorf("%s", err)
				break
			}

			if time.Now().After(lastAccessTime.Add(LIMIT_TIME)) {
				actualService := etcdcron.services[serviceName].Get(service.index)
				if actualService != nil {
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
