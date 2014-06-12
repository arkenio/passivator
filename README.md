

Passivator
======

[![Build Status](https://travis-ci.org/arkenio/passivator.png?branch=master)](https://travis-ci.org/arkenio/passivator)

The passivator process purpose is watching last access services dates submitted by [Gogeta](http://github.com/arkenio/gogeta) and starting/stopping those services depending on limit access duration (If no activity has been detected during 12 hours, service is passivated). The process will restart after first new activity detected.

It is part of the nuxeo.io infrastructure.

How it works
-------------

The passivator is:

* watching etcd services entries to re-activate service when a user is trying to access it
* executing a cron process to check every 5 minutes if some services need to be passivated after 12 hours without activity

Configuration
-------------

Several parameters allow to configure the way the passivator behaves :

 * `serviceDir` allows to select the prefix of the key where it watches for services (default value: "/services" )
 * `etcdAddress` specifies the address of the `etcd` server (default value: "http://172.17.42.1:4001")
 * `cronDuration` specifies the lap duration in minutes to check services to passivate (default value: "5")
 * `passiveLimitDuration` specifies the limit duration in hours when a service has to be passivated if no activity has been detected (default value: "12")
 
Example:
 
         passivator -etcdAddress="http://172.17.42.1:4001" \
               -serviceDir="/services" \
               -cronDuration="5"
               -passiveLimitDuration="12"

If you need to display Info logs this flag can be added:

               -stderrthreshold=INFO


Report & Contribute
-------------------

We are glad to welcome new developers on this initiative, and even simple usage feedback is great.
- Ask your questions on [Nuxeo Answers](http://answers.nuxeo.com)
- Report issues on this github repository (see [issues link](http://github.com/arkenio/passivator/issues) on the right)
- Contribute: Send pull requests!


About Nuxeo
-----------

Nuxeo provides a modular, extensible Java-based
[open source software platform for enterprise content management](http://www.nuxeo.com/en/products/ep),
and packaged applications for [document management](http://www.nuxeo.com/en/products/document-management),
[digital asset management](http://www.nuxeo.com/en/products/dam) and
[case management](http://www.nuxeo.com/en/products/case-management).

Designed by developers for developers, the Nuxeo platform offers a modern
architecture, a powerful plug-in model and extensive packaging
capabilities for building content applications.

More information on: <http://www.nuxeo.com/>
