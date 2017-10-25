// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
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

var _ = Describe("Disk", func() {
	var (
		server     *mocks.Server
		client     *Client
		tenantID   string
		projID     string
		flavorName string
		flavorID   string
		diskSpec   *DiskCreateSpec
	)

	BeforeEach(func() {
		server, client = testSetup()
		tenantID = createTenant(server, client)
		projID = createProject(server, client, tenantID)
		flavorName, flavorID = createFlavor(server, client)
		diskSpec = &DiskCreateSpec{
			Flavor:     flavorName,
			Kind:       "persistent-disk",
			CapacityGB: 2,
			Name:       randomString(10, "go-sdk-disk-"),
		}
	})

	AfterEach(func() {
		cleanDisks(client, projID)
		cleanFlavors(client)
		cleanTenants(client)
		server.Close()
	})

	Describe("CreateAndDeleteDisk", func() {
		It("Disk create and delete succeeds", func() {
			mockTask := createMockTask("CREATE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.CreateDisk(projID, diskSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_DISK"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("DELETE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Disks.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_DISK"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("GetDisk", func() {
		It("Get disk returns a disk ID", func() {
			mockTask := createMockTask("CREATE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.CreateDisk(projID, diskSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			diskMock := &PersistentDisk{
				Name:       diskSpec.Name,
				Flavor:     diskSpec.Flavor,
				CapacityGB: diskSpec.CapacityGB,
				Kind:       diskSpec.Kind,
			}
			server.SetResponseJson(200, diskMock)
			disk, err := client.Disks.Get(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(disk.Name).Should(Equal(diskSpec.Name))
			Expect(disk.Flavor).Should(Equal(diskSpec.Flavor))
			Expect(disk.Kind).Should(Equal(diskSpec.Kind))
			Expect(disk.CapacityGB).Should(Equal(diskSpec.CapacityGB))
			Expect(disk.ID).Should(Equal(task.Entity.ID))

			mockTask = createMockTask("DELETE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Disks.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("GetTasks", func() {
		It("GetTasks returns a completed task", func() {
			mockTask := createMockTask("CREATE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.CreateDisk(projID, diskSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			mockTasksPage := createMockTasksPage(*mockTask)
			server.SetResponseJson(200, mockTasksPage)
			taskList, err := client.Disks.GetTasks(task.Entity.ID, &TaskGetOptions{})

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(taskList).ShouldNot(BeNil())
			Expect(taskList.Items).Should(ContainElement(*task))

			mockTask = createMockTask("DELETE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Disks.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("ManageDiskIamPolicy", func() {
		It("Set IAM Policy succeeds", func() {
			mockTask := createMockTask("CREATE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.CreateDisk(projID, diskSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_DISK"))
			Expect(task.State).Should(Equal("COMPLETED"))

			diskID := task.Entity.ID
			mockTask = createMockTask("SET_IAM_POLICY", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			var policy []*RoleBinding
			policy = []*RoleBinding{{Role: "owner", Subjects: []string{"joe@photon.local"}}}
			task, err = client.Disks.SetIam(diskID, policy)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("SET_IAM_POLICY"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("DELETE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Disks.Delete(diskID)
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_DISK"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Modify IAM Policy succeeds", func() {
			mockTask := createMockTask("CREATE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.CreateDisk(projID, diskSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_DISK"))
			Expect(task.State).Should(Equal("COMPLETED"))

			diskID := task.Entity.ID
			mockTask = createMockTask("MODIFY_IAM_POLICY", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			var delta []*RoleBindingDelta
			delta = []*RoleBindingDelta{{Subject: "joe@photon.local", Action: "ADD", Role: "owner"}}
			task, err = client.Disks.ModifyIam(diskID, delta)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("MODIFY_IAM_POLICY"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("DELETE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Disks.Delete(diskID)
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_DISK"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Get IAM Policy succeeds", func() {
			mockTask := createMockTask("CREATE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.CreateDisk(projID, diskSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_DISK"))
			Expect(task.State).Should(Equal("COMPLETED"))

			diskID := task.Entity.ID
			mockTask = createMockTask("SET_IAM_POLICY", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			var policy []*RoleBinding
			policy = []*RoleBinding{{Role: "owner", Subjects: []string{"joe@photon.local"}}}
			task, err = client.Disks.SetIam(diskID, policy)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("SET_IAM_POLICY"))
			Expect(task.State).Should(Equal("COMPLETED"))

			server.SetResponseJson(200, policy)
			response, err := client.Disks.GetIam(diskID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(response[0].Subjects).Should(Equal(policy[0].Subjects))
			Expect(response[0].Role).Should(Equal(policy[0].Role))

			mockTask = createMockTask("DELETE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Disks.Delete(diskID)
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_DISK"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})
})
