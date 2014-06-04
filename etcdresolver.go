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

const (
	SERVICE_DOMAINTYTPE = "service"
	URI_DOMAINTYPE      = "uri"
	TIME_FORMAT         = "2006-01-02 15:04:05"
)

type Service struct {
	index      string
	nodeKey    string
	domain     string
	name       string
	lastAccess string
}

type EtcdResolver struct {
	config          *Config
	watcher         *watcher
	services        map[string]*ServiceCluster
	watchIndex      uint64
	etcdcron        *EtcdCron
}

func NewEtcdResolver(config *Config) (*EtcdResolver, error) {
	services := make(map[string]*ServiceCluster)
	watcher, error := NewEtcdWatcher(config, services)
	if error != nil {
		return nil, error
	}
	etcdcron, error := NewEtcdCron(config, services)
	if error != nil {
		return nil, error
	}
	return &EtcdResolver{config, watcher, services, 0, etcdcron}, nil
}

func (resolver *EtcdResolver) init() {
	go resolver.etcdcron.init()
	resolver.watcher.init()
}
