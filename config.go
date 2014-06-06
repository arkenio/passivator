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
	"errors"
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	"strings"
	"regexp"
	"flag"
)

type Config struct {
	servicePrefix        string
	etcdAddress          string
	cronDuration         string
	passiveLimitDuration string
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

func (c *Config) getServiceForNode(node *etcd.Node, config *Config) string {
	r := regexp.MustCompile(config.servicePrefix + "/(.*)(/.*)*")
	return strings.Split(r.FindStringSubmatch(node.Key)[1], "/")[0]
}

func (c *Config) getServiceIndexForNode(node *etcd.Node, config *Config) string {
	r := regexp.MustCompile(config.servicePrefix + "/(.*)(/.*)*")
	return strings.Split(r.FindStringSubmatch(node.Key)[1], "/")[1]
}

func parseConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.servicePrefix, "serviceDir", "/services", "etcd prefix to get services")
	flag.StringVar(&config.etcdAddress, "etcdAddress", "http://127.0.0.1:4001/", "etcd client host")
	flag.StringVar(&config.cronDuration, "cronDuration", "5", "Passivation cron checking duration ")
	flag.StringVar(&config.passiveLimitDuration, "passiveLimitDuration", "12", "Limit duration of passivation")
	flag.Parse()

	glog.Infof("Dumping Configuration")
	glog.Infof("  servicesPrefix : %s", config.servicePrefix)
	glog.Infof("  etcdAddress : %s", config.etcdAddress)
	glog.Infof("  Passivation cron duration: %s", config.cronDuration)
	glog.Infof("  Limit duration of passivation : %s", config.passiveLimitDuration)

	return config
}

