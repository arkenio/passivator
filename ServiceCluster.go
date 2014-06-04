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
	"sync"
	"github.com/golang/glog"
)

type ServiceCluster struct {
	instances []*Service
	lastIndex int
	lock      sync.RWMutex
}

func (cl *ServiceCluster) Remove(instanceIndex string) {

	match := -1
	for k, v := range cl.instances {
		if v.index == instanceIndex {
			match = k
		}
	}

	cl.instances = append(cl.instances[:match], cl.instances[match+1:]...)
	cl.Dump("remove")
}

// Get an service by its key (index). Returns nil if not found.
func (cl *ServiceCluster) Get(instanceIndex string) *Service {
	for i, v := range cl.instances {
		if v.index == instanceIndex {
			return cl.instances[i]
		}
	}
	return nil
}

func (cl *ServiceCluster) Add(service *Service) {
	for index, v := range cl.instances {
		if v.index == service.index {
			cl.instances[index] = service
			return
		}
	}

	cl.instances = append(cl.instances, service)
}

func (cl *ServiceCluster) Dump(action string) {
	for _, v := range cl.instances {
		glog.Infof("Dump after %s %s", action, v.index)
	}
}
