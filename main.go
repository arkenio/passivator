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

