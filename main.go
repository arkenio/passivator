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

const (
	progName = "Passivator"
)

func main() {
	glog.Infof("%s starting", progName)

	config := parseConfig()

	resolver, err := NewEtcdResolver(config)
	if err == nil {
		resolver.init()
	}
}
