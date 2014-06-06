package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
	"strconv"
)

func test_passivator(t *testing.T) {
	var service *Service
	var etcdcron *EtcdCron
	var watcher *watcher
	config := parseConfig()

	Convey("Given a last access and service status", t, func() {

			Convey("When last access is too old", func() {
					service.status.expected = "started"
					service.status.current = "started"
					lastAccessTime, _ := time.Parse(TIME_FORMAT, "1984-02-29 15:00:00")
					service.lastAccess = &lastAccessTime
					parameter, _ := strconv.Atoi(config.passiveLimitDuration)
					passiveLimitDuration := time.Duration(parameter) * time.Hour

					Convey("Then the service should be passivated", func() {
							So(etcdcron.hasToBePassivated(service, passiveLimitDuration), ShouldEqual, true)
						})

				})

			Convey("When last access is recent and service passivated", func() {
					service.status.expected = "passivated"
					service.status.current = "passivated"
					lastAccessTime, _ := time.Parse(TIME_FORMAT, time.Now().String())
					service.lastAccess = &lastAccessTime
					parameter, _ := strconv.Atoi(config.passiveLimitDuration)
					passiveLimitDuration := time.Duration(parameter) * time.Hour

					Convey("Then the service should be restarted", func() {
							So(watcher.hasToBeActivated(service, passiveLimitDuration), ShouldEqual, true)
						})

				})
		})
}

