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

var _ = Describe("Router", func() {
	var (
		server           *mocks.Server
		client           *Client
		routerCreateSpec *RouterCreateSpec
		tenantID         string
		projID           string
	)

	BeforeEach(func() {
		server, client = testSetup()
		tenantID = createTenant(server, client)
		projID = createProject(server, client, tenantID)
		routerCreateSpec = &RouterCreateSpec{Name: "router-1", PrivateIpCidr: "cidr1"}
	})

	AfterEach(func() {
		cleanTenants(client)
		server.Close()
	})

	Describe("CreateDeleteRouter", func() {
		It("Router create and delete succeeds", func() {
			mockTask := createMockTask("CREATE_ROUTER", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.CreateRouter(projID, routerCreateSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_ROUTER"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("DELETE_ROUTER", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Routers.Delete("routerId")
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_ROUTER"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("GetRouter", func() {
		It("Get returns router", func() {
			mockTask := createMockTask("CREATE_ROUTER", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.CreateRouter(projID, routerCreateSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_ROUTER"))
			Expect(task.State).Should(Equal("COMPLETED"))

			server.SetResponseJson(200, Router{Name: "router-1", PrivateIpCidr: "cidr1"})
			router, err := client.Routers.Get(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(router).ShouldNot(BeNil())

			var found bool
			if router.Name == "router-1" && router.ID == task.Entity.ID {
				found = true
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_ROUTER", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			_, err = client.Routers.Delete(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("UpdateRouter", func() {
		It("update router's name", func() {
			mockTask := createMockTask("UPDATE_ROUTER", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			routerSpec := &RouterUpdateSpec{RouterName: "router-1"}
			task, err := client.Routers.UpdateRouter("router-Id", routerSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("UPDATE_ROUTER"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("GetRouterSubnets", func() {
		It("GetAll returns subnet", func() {
			mockTask := createMockTask("CREATE_SUBNET", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			subnetSpec := &SubnetCreateSpec{
				Name:          "subnet_name",
				Description:   "subnet description",
				PrivateIpCidr: "192.168.0.1/24",
			}

			task, err := client.Routers.CreateSubnet("router-id", subnetSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			subnetMock := Subnet{
				Name:          subnetSpec.Name,
				Description:   subnetSpec.Description,
				PrivateIpCidr: subnetSpec.PrivateIpCidr,
			}
			server.SetResponseJson(200, &Subnets{[]Subnet{subnetMock}})
			subnetList, err := client.Routers.GetSubnets("router-id", &SubnetGetOptions{})
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(subnetList).ShouldNot(BeNil())

			var found bool
			for _, subnet := range subnetList.Items {
				if subnet.Name == subnetSpec.Name && subnet.ID == task.Entity.ID {
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
	})

})
