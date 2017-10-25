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

var _ = Describe("Subnet", func() {
	var (
		server                   *mocks.Server
		client                   *Client
		subnetCreateSpec         *SubnetCreateSpec
		subnetSpecWithPortGroups *SubnetCreateSpec
		portGroups               *PortGroups
		tenantID                 string
		projID                   string
		routerID                 string
	)

	BeforeEach(func() {
		server, client = testSetup()
		tenantID = createTenant(server, client)
		projID = createProject(server, client, tenantID)
		routerID = createRouter(server, client, projID)
		portGroups = &PortGroups{Names: []string{"portGroup"}}
		subnetCreateSpec = &SubnetCreateSpec{Name: "subnet-1", Description: "Test subnet", PrivateIpCidr: "cidr1"}
		subnetSpecWithPortGroups = &SubnetCreateSpec{Name: "subnet-1", Description: "Test subnet", PrivateIpCidr: "cidr1"}
	})

	AfterEach(func() {
		cleanSubnets(client)
		cleanTenants(client)
		server.Close()
	})

	Describe("CreateDeleteSubnet", func() {
		It("Subnet create and delete succeeds", func() {
			mockTask := createMockTask("CREATE_SUBNET", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Routers.CreateSubnet(routerID, subnetCreateSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_SUBNET"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("DELETE_SUBNET", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Subnets.Delete("subnet-Id")
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_SUBNET"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("CreateDeletePortGroup", func() {
		It("PortGroup create and delete succeeds", func() {
			mockTask := createMockTask("CREATE_PORT_GROUP", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Subnets.Create(subnetSpecWithPortGroups)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_PORT_GROUP"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("DELETE_PORT_GROUP", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Subnets.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_PORT_GROUP"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("GetSubnet", func() {
		It("Get returns subnet", func() {
			mockTask := createMockTask("CREATE_SUBNET", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Routers.CreateSubnet(routerID, subnetCreateSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_SUBNET"))
			Expect(task.State).Should(Equal("COMPLETED"))

			server.SetResponseJson(200, Subnet{Name: "subnet-1", Description: "Test subnet", PrivateIpCidr: "cidr1"})
			subnet, err := client.Subnets.Get(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(subnet).ShouldNot(BeNil())

			var found bool
			if subnet.Name == "subnet-1" && subnet.ID == task.Entity.ID {
				found = true
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_SUBNET", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			_, err = client.Subnets.Delete(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})

		It("GetAll Subnet succeeds", func() {
			mockTask := createMockTask("CREATE_SUBNET", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err := client.Subnets.Create(subnetSpecWithPortGroups)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			server.SetResponseJson(200, createMockSubnetsPage(Subnet{Name: subnetSpecWithPortGroups.Name}))
			subnets, err := client.Subnets.GetAll(&SubnetGetOptions{})

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(subnets).ShouldNot(BeNil())

			var found bool
			for _, subnet := range subnets.Items {
				if subnet.Name == subnetSpecWithPortGroups.Name && subnet.ID == task.Entity.ID {
					found = true
					break
				}
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_SUBNET", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Subnets.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})

		It("GetAll PortGroups with optional name succeeds", func() {
			mockTask := createMockTask("CREATE_PORT_GROUP", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err := client.Subnets.Create(subnetSpecWithPortGroups)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			server.SetResponseJson(200, createMockSubnetsPage(Subnet{Name: subnetSpecWithPortGroups.Name}))
			subnets, err := client.Subnets.GetAll(&SubnetGetOptions{Name: subnetSpecWithPortGroups.Name})

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(subnets).ShouldNot(BeNil())

			var found bool
			for _, subnet := range subnets.Items {
				if subnet.Name == subnetSpecWithPortGroups.Name && subnet.ID == task.Entity.ID {
					found = true
					break
				}
			}
			Expect(len(subnets.Items)).Should(Equal(1))
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_PORT_GROUP", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Subnets.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("UpdateSubnet", func() {
		It("update subnet's name", func() {
			mockTask := createMockTask("UPDATE_SUBNET", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			subnetSpec := &SubnetUpdateSpec{SubnetName: "subnet-1"}
			task, err := client.Subnets.Update("subnet-Id", subnetSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("UPDATE_SUBNET"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("SetDefaultSubnet", func() {
		It("Set default subnet succeeds", func() {
			mockTask := createMockTask("SET_DEFAULT_SUBNET", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Subnets.SetDefault("subnetId")
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("SET_DEFAULT_SUBNET"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})
})
