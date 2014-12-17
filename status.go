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

const (
	STARTING_STATUS   = "starting"
	STARTED_STATUS    = "started"
	STOPPING_STATUS   = "stopping"
	STOPPED_STATUS    = "stopped"
	ERROR_STATUS      = "error"
	NA_STATUS         = "n/a"
	PASSIVATED_STATUS = "passivated"
)

type Status struct {
	alive    string
	current  string
	expected string
}
