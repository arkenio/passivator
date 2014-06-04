

Passivator
======

The passivator process purpose is watching last access services dates and starting/stopping those depending on the duration. When an access is detected on passivated service, the passivator will re-starting it.

It is part of the nuxeo.io infrastructure.

Configuration
-------------

Several parameters allow to configure the way the passivator behaves :

 * `serviceDir` allows to select the prefix of the key where it watches for services
 * `etcdAddress` specify the address of the `etcd` server

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
