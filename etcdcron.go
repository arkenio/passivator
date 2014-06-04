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
	ticker := updateTicker()
	for {
		<-ticker.C
		// TODO Check every 5 minutes all services lastAccess etcd date entry
		ticker = updateTicker()
	}
}

func updateTicker() *time.Ticker {
	return time.NewTicker(INTERVAL_PERIOD)
}
