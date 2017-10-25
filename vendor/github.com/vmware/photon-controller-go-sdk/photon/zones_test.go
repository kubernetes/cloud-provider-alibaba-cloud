// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package photon

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/photon-controller-go-sdk/photon/internal/mocks"
)

var _ = Describe("Zone", func() {
	var (
		server *mocks.Server
		client *Client
	)

	BeforeEach(func() {
		server, client = testSetup()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("CreateAndDeleteZone", func() {
		It("Zone create and delete succeeds", func() {
			mockTask := createMockTask("CREATE_ZONE", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			zoneSpec := &ZoneCreateSpec{Name: randomString(10, "go-sdk-zone-")}
			task, err := client.Zones.Create(zoneSpec)
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_ZONE"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("DELETE_ZONE", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Zones.Delete(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_ZONE"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Zone create fails", func() {
			zoneSpec := &ZoneCreateSpec{}
			task, err := client.Zones.Create(zoneSpec)

			Expect(err).ShouldNot(BeNil())
			Expect(task).Should(BeNil())
		})
	})

	Describe("GetZone", func() {
		It("Get returns zone", func() {
			mockTask := createMockTask("CREATE_ZONE", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			zoneName := randomString(10, "go-sdk-zone-")
			zoneSpec := &ZoneCreateSpec{Name: zoneName}
			task, err := client.Zones.Create(zoneSpec)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_ZONE"))
			Expect(task.State).Should(Equal("COMPLETED"))

			server.SetResponseJson(200, Zone{Name: zoneName})
			zone, err := client.Zones.Get(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(zone).ShouldNot(BeNil())

			var found bool
			if zone.Name == zoneName && zone.ID == task.Entity.ID {
				found = true
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_ZONE", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			_, err = client.Zones.Delete(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})

		It("Get all returns zones", func() {
			mockTask := createMockTask("CREATE_ZONE", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			zoneName := randomString(10, "go-sdk-zone-")
			zoneSpec := &ZoneCreateSpec{Name: zoneName}
			task, err := client.Zones.Create(zoneSpec)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_ZONE"))
			Expect(task.State).Should(Equal("COMPLETED"))

			zonePage := createMockZonesPage(Zone{Name: zoneName})
			server.SetResponseJson(200, zonePage)
			zones, err := client.Zones.GetAll()

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(zones).ShouldNot(BeNil())

			var found bool
			for _, zone := range zones.Items {
				if zone.Name == zoneName && zone.ID == task.Entity.ID {
					found = true
					break
				}
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_ZONE", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			_, err = client.Zones.Delete(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("GetZoneTasks", func() {
		var (
			option string
		)

		Context("no extra options for GetTask", func() {
			BeforeEach(func() {
				option = ""
			})

			It("GetTasks returns a completed task", func() {
				mockTask := createMockTask("CREATE_ZONE", "COMPLETED")
				mockTask.Entity.ID = "mock-task-id"
				server.SetResponseJson(200, mockTask)
				zoneSpec := &ZoneCreateSpec{Name: randomString(10, "go-sdk-zone-")}
				task, err := client.Zones.Create(zoneSpec)

				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
				Expect(task).ShouldNot(BeNil())
				Expect(task.Operation).Should(Equal("CREATE_ZONE"))
				Expect(task.State).Should(Equal("COMPLETED"))

				mockTasksPage := createMockTasksPage(*mockTask)
				server.SetResponseJson(200, mockTasksPage)
				taskList, err := client.Zones.GetTasks(task.Entity.ID, &TaskGetOptions{State: option})
				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
				Expect(taskList).ShouldNot(BeNil())
				Expect(taskList.Items).Should(ContainElement(*task))

				mockTask = createMockTask("DELETE_ZONE", "COMPLETED")
				server.SetResponseJson(200, mockTask)
				_, err = client.Zones.Delete(task.Entity.ID)

				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
			})
		})

		Context("Searching COMPLETED state for GetTask", func() {
			BeforeEach(func() {
				option = "COMPLETED"
			})

			It("GetTasks returns a completed task", func() {
				mockTask := createMockTask("CREATE_ZONE", "COMPLETED")
				mockTask.Entity.ID = "mock-task-id"
				server.SetResponseJson(200, mockTask)
				zoneSpec := &ZoneCreateSpec{Name: randomString(10, "go-sdk-zone-")}
				task, err := client.Zones.Create(zoneSpec)

				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
				Expect(task).ShouldNot(BeNil())
				Expect(task.Operation).Should(Equal("CREATE_ZONE"))
				Expect(task.State).Should(Equal("COMPLETED"))

				mockTasksPage := createMockTasksPage(*mockTask)
				server.SetResponseJson(200, mockTasksPage)
				taskList, err := client.Zones.GetTasks(task.Entity.ID, &TaskGetOptions{State: option})
				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
				Expect(taskList).ShouldNot(BeNil())
				Expect(taskList.Items).Should(ContainElement(*task))

				mockTask = createMockTask("DELETE_ZONE", "COMPLETED")
				server.SetResponseJson(200, mockTask)
				_, err = client.Zones.Delete(task.Entity.ID)

				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
			})
		})
	})
})
