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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/photon-controller-go-sdk/photon/internal/mocks"
	"reflect"
)

var _ = Describe("Project", func() {
	var (
		server     *mocks.Server
		client     *Client
		tenantID   string
		projID     string
		flavorName string
		flavorID   string
	)

	BeforeEach(func() {
		server, client = testSetup()
		tenantID = createTenant(server, client)
		projID = createProject(server, client, tenantID)
		flavorName, flavorID = createFlavor(server, client)

	})

	AfterEach(func() {
		cleanDisks(client, projID)
		cleanFlavors(client)
		cleanTenants(client)
		server.Close()
	})

	Describe("GetProjectTasks", func() {
		It("GetTasks returns a completed task", func() {
			mockTask := createMockTask("CREATE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			diskSpec := &DiskCreateSpec{
				Flavor:     flavorName,
				Kind:       "persistent-disk",
				CapacityGB: 2,
				Name:       randomString(10, "go-sdk-disk-"),
			}

			task, err := client.Projects.CreateDisk(projID, diskSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			mockTasksPage := createMockTasksPage(*mockTask)
			server.SetResponseJson(200, mockTasksPage)
			taskList, err := client.Projects.GetTasks(projID, &TaskGetOptions{})
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(taskList).ShouldNot(BeNil())
			Expect(taskList.Items).Should(ContainElement(*task))

			// Clean disk
			mockTask = createMockTask("DELETE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Disks.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("GetProjectDisks", func() {
		It("GetAll returns disk", func() {
			mockTask := createMockTask("CREATE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			diskSpec := &DiskCreateSpec{
				Flavor:     flavorName,
				Kind:       "persistent-disk",
				CapacityGB: 2,
				Name:       randomString(10, "go-sdk-disk-"),
			}

			task, err := client.Projects.CreateDisk(projID, diskSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			diskMock := PersistentDisk{
				Name:       diskSpec.Name,
				Flavor:     diskSpec.Flavor,
				CapacityGB: diskSpec.CapacityGB,
				Kind:       diskSpec.Kind,
			}
			server.SetResponseJson(200, &DiskList{[]PersistentDisk{diskMock}})
			diskList, err := client.Projects.GetDisks(projID, &DiskGetOptions{})
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(diskList).ShouldNot(BeNil())

			var found bool
			for _, disk := range diskList.Items {
				if disk.Name == diskSpec.Name && disk.ID == task.Entity.ID {
					found = true
					break
				}
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_DISK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Disks.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("GetProjectVms", func() {
		var (
			imageID      string
			flavorSpec   *FlavorCreateSpec
			vmFlavorSpec *FlavorCreateSpec
		)

		BeforeEach(func() {
			imageID = createImage(server, client)
			flavorSpec = &FlavorCreateSpec{
				[]QuotaLineItem{QuotaLineItem{"COUNT", 1, "ephemeral-disk.cost"}},
				"ephemeral-disk",
				randomString(10, "go-sdk-flavor-"),
			}

			_, err := client.Flavors.Create(flavorSpec)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			vmFlavorSpec = &FlavorCreateSpec{
				Name: randomString(10, "go-sdk-flavor-"),
				Kind: "vm",
				Cost: []QuotaLineItem{
					QuotaLineItem{"GB", 2, "vm.memory"},
					QuotaLineItem{"COUNT", 4, "vm.cpu"},
				},
			}
			_, err = client.Flavors.Create(vmFlavorSpec)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})

		AfterEach(func() {
			cleanVMs(client, projID)
		})

		It("GetAll returns vm", func() {
			mockTask := createMockTask("CREATE_VM", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			vmSpec := &VmCreateSpec{
				Flavor:        vmFlavorSpec.Name,
				SourceImageID: imageID,
				AttachedDisks: []AttachedDisk{
					AttachedDisk{
						CapacityGB: 1,
						Flavor:     flavorSpec.Name,
						Kind:       "ephemeral-disk",
						Name:       randomString(10),
						State:      "STARTED",
						BootDisk:   true,
					},
				},
				Name: randomString(10, "go-sdk-vm-"),
			}

			task, err := client.Projects.CreateVM(projID, vmSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			mockVm := VM{Name: vmSpec.Name}
			server.SetResponseJson(200, createMockVmsPage(mockVm))
			vmList, err := client.Projects.GetVMs(projID, &VmGetOptions{})
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(vmList).ShouldNot(BeNil())

			var found bool
			for _, vm := range vmList.Items {
				if vm.Name == vmSpec.Name && vm.ID == task.Entity.ID {
					found = true
					break
				}
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_VM", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.VMs.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("GetProjectRouters", func() {
		It("GetAll returns router", func() {
			mockTask := createMockTask("CREATE_ROUTER", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			routerSpec := &RouterCreateSpec{
				Name:          "router_name",
				PrivateIpCidr: "192.168.0.1/24",
			}

			task, err := client.Projects.CreateRouter(projID, routerSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			routerMock := Router{
				Name:          routerSpec.Name,
				PrivateIpCidr: routerSpec.PrivateIpCidr,
			}
			server.SetResponseJson(200, &Routers{[]Router{routerMock}})
			routerList, err := client.Projects.GetRouters(projID, &RouterGetOptions{})
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(routerList).ShouldNot(BeNil())

			var found bool
			for _, router := range routerList.Items {
				if router.Name == routerSpec.Name && router.ID == task.Entity.ID {
					found = true
					break
				}
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_ROUTER", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Routers.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("GetProjectNetworks", func() {
		It("GetAll returns network", func() {
			mockTask := createMockTask("CREATE_NETWORK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			networkSpec := &NetworkCreateSpec{
				Name:          "network_name",
				PrivateIpCidr: "192.168.0.1/24",
			}

			task, err := client.Projects.CreateNetwork(projID, networkSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			networkMock := Network{
				Name:          networkSpec.Name,
				PrivateIpCidr: networkSpec.PrivateIpCidr,
			}
			server.SetResponseJson(200, &Networks{[]Network{networkMock}})
			networkList, err := client.Projects.GetNetworks(projID, &NetworkGetOptions{})
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(networkList).ShouldNot(BeNil())

			var found bool
			for _, network := range networkList.Items {
				if network.Name == networkSpec.Name && network.ID == task.Entity.ID {
					found = true
					break
				}
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_NETWORK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Networks.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("GetProjectServices", func() {
		It("GetAll returns service", func() {
			if isIntegrationTest() {
				Skip("Skipping service test on integration mode. Need to set extendedProperties to use real IPs and masks")
			}
			mockTask := createMockTask("CREATE_SERVICE", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			serviceSpec := &ServiceCreateSpec{
				Name:               randomString(10, "go-sdk-service-"),
				Type:               "KUBERNETES",
				WorkerCount:        50,
				BatchSizeWorker:    5,
				ExtendedProperties: map[string]string{},
			}

			task, err := client.Projects.CreateService(projID, serviceSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			mockService := Service{Name: serviceSpec.Name}
			server.SetResponseJson(200, createMockServicesPage(mockService))
			serviceList, err := client.Projects.GetServices(projID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(serviceList).ShouldNot(BeNil())

			var found bool
			for _, service := range serviceList.Items {
				if service.Name == serviceSpec.Name && service.ID == task.Entity.ID {
					found = true
					break
				}
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_SERVICE", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Services.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("SecurityGroups", func() {
		It("sets security groups for a project", func() {
			// Set security groups for the project
			expected := &Tenant{
				SecurityGroups: []SecurityGroup{
					{randomString(10), false},
					{randomString(10), false},
				},
			}
			mockTask := createMockTask("SET_TENANT_SECURITY_GROUPS", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			securityGroups := &SecurityGroupsSpec{
				[]string{expected.SecurityGroups[0].Name, expected.SecurityGroups[1].Name},
			}
			createTask, err := client.Projects.SetSecurityGroups(projID, securityGroups)
			createTask, err = client.Tasks.Wait(createTask.ID)
			Expect(err).Should(BeNil())

			// Get the security groups for the project
			server.SetResponseJson(200, expected)
			project, err := client.Projects.Get(projID)
			Expect(err).Should(BeNil())
			fmt.Fprintf(GinkgoWriter, "Got project: %+v", project)
			Expect(expected.SecurityGroups).Should(Equal(project.SecurityGroups))
		})
	})

	Describe("ProjectQuota", func() {

		It("Get Project Quota succeeds", func() {
			mockQuota := createMockQuota()

			// Get current Quota
			server.SetResponseJson(200, mockQuota)
			quota, err := client.Projects.GetQuota(tenantID)

			GinkgoT().Log(err)
			eq := reflect.DeepEqual(quota.QuotaLineItems, mockQuota.QuotaLineItems)
			Expect(eq).Should(Equal(true))
		})

		It("Set Project Quota succeeds", func() {
			mockQuotaSpec := &QuotaSpec{
				"vmCpu":        {Unit: "COUNT", Limit: 10, Usage: 0},
				"vmMemory":     {Unit: "GB", Limit: 18, Usage: 0},
				"diskCapacity": {Unit: "GB", Limit: 100, Usage: 0},
			}

			mockTask := createMockTask("SET_QUOTA", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.SetQuota(tenantID, mockQuotaSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("SET_QUOTA"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Update Project Quota succeeds", func() {
			mockQuotaSpec := &QuotaSpec{
				"vmCpu":        {Unit: "COUNT", Limit: 30, Usage: 0},
				"vmMemory":     {Unit: "GB", Limit: 40, Usage: 0},
				"diskCapacity": {Unit: "GB", Limit: 150, Usage: 0},
			}

			mockTask := createMockTask("UPDATE_QUOTA", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.UpdateQuota(tenantID, mockQuotaSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("UPDATE_QUOTA"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Exclude Project Quota Items succeeds", func() {
			mockQuotaSpec := &QuotaSpec{
				"vmCpu2":    {Unit: "COUNT", Limit: 10, Usage: 0},
				"vmMemory3": {Unit: "GB", Limit: 18, Usage: 0},
			}

			mockTask := createMockTask("DELETE_QUOTA", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.ExcludeQuota(tenantID, mockQuotaSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_QUOTA"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})
})

var _ = Describe("IAM", func() {
	var (
		server   *mocks.Server
		client   *Client
		tenantID string
		projID   string
	)

	BeforeEach(func() {
		server, client = testSetup()
		tenantID = createTenant(server, client)
		projID = createProject(server, client, tenantID)
	})

	AfterEach(func() {
		cleanTenants(client)
		server.Close()
	})

	Describe("ManageProjectIamPolicy", func() {
		It("Set IAM Policy succeeds", func() {
			mockTask := createMockTask("SET_IAM_POLICY", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			var policy []*RoleBinding
			policy = []*RoleBinding{{Role: "owner", Subjects: []string{"joe@photon.local"}}}
			task, err := client.Projects.SetIam(projID, policy)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("SET_IAM_POLICY"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Modify IAM Policy succeeds", func() {
			mockTask := createMockTask("MODIFY_IAM_POLICY", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			var delta []*RoleBindingDelta
			delta = []*RoleBindingDelta{{Subject: "joe@photon.local", Action: "ADD", Role: "owner"}}
			task, err := client.Projects.ModifyIam(projID, delta)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("MODIFY_IAM_POLICY"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Get IAM Policy succeeds", func() {
			var policy []*RoleBinding
			policy = []*RoleBinding{{Role: "owner", Subjects: []string{"joe@photon.local"}}}
			server.SetResponseJson(200, policy)
			response, err := client.Projects.GetIam(projID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(response[0].Subjects).Should(Equal(policy[0].Subjects))
			Expect(response[0].Role).Should(Equal(policy[0].Role))
		})
	})
})
