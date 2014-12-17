package main

import (
	"github.com/arkenio/goarken"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func Test_passivator(t *testing.T) {
	var service *goarken.Service

	config := parseConfig()
	var p *Passivator

	var a *Activator

	Convey("Given a last access and service status", t, func() {
		service = &goarken.Service{}
		status := &goarken.Status{Service: service}
		service.Status = status
		lastAccessTime, _ := time.Parse(goarken.TIME_FORMAT, "1984-02-29 15:00:00")
		service.LastAccess = &lastAccessTime

		p = &Passivator{Config: config}
		a = &Activator{Config: config}

		Convey("When last access is too old", func() {
			service.Status.Expected = goarken.STARTED_STATUS
			service.Status.Current = goarken.STARTED_STATUS

			Convey("Then the service should be passivated", func() {
				So(p.hasToBePassivated(service), ShouldEqual, true)
			})

		})

		Convey("When last access is recent and service passivated", func() {
			service.Status.Expected = goarken.STARTED_STATUS
			service.Status.Current = goarken.STOPPED_STATUS

			Convey("Then the service should be restarted", func() {
				So(a.hasToBeRestarted(service), ShouldEqual, true)
			})

		})
	})
}
