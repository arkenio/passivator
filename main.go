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

import "github.com/golang/glog"
import (
	"github.com/arkenio/goarken"
	"os"
	"os/signal"
	"syscall"
)

const (
	progName = "Passivator"
)

func main() {
	glog.Infof("%s starting", progName)

	config := parseConfig()

	domains := make(map[string]*goarken.Domain)
	services := make(map[string]*goarken.ServiceCluster)

	client, err := config.getEtcdClient()

	if err != nil {
		panic(err)
	}

	w := &goarken.Watcher{
		Client:        client,
		DomainPrefix:  "/domains",
		ServicePrefix: "/services",
		Domains:       domains,
		Services:      services,
	}

	stop := goarken.NewBroadcaster()

	go NewPassivator(w, config, stop.Listen()).run()
	go NewActivator(w, config, stop.Listen()).run()

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	glog.Info(<-ch)
	stop.Write(struct{}{})

}
