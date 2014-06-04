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
	"errors"
	"flag"
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
)

type Config struct {
	servicePrefix string
	etcdAddress   string
	client        *etcd.Client
}

func (c *Config) getEtcdClient() (*etcd.Client, error) {
	if c.client == nil {
		c.client = etcd.NewClient([]string{c.etcdAddress})
		if !c.client.SyncCluster() {
			return nil, errors.New("Unable to sync with etcd cluster, check your configuration or etcd status")
		}
	}
	return c.client, nil
}

func parseConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.servicePrefix, "serviceDir", "/services", "etcd prefix to get services")
	flag.StringVar(&config.etcdAddress, "etcdAddress", "http://127.0.0.1:4001/", "etcd client host")
	flag.Parse()

	glog.Infof("Dumping Configuration")
	glog.Infof("  servicesPrefix : %s", config.servicePrefix)
	glog.Infof("  etcdAddress : %s", config.etcdAddress)

	return config
}
