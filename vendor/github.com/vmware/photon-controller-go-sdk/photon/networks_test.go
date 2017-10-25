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

var _ = Describe("Network", func() {
	var (
		server            *mocks.Server
		client            *Client
		networkCreateSpec *NetworkCreateSpec
		tenantID          string
		projID            string
	)

	BeforeEach(func() {
		server, client = testSetup()
		tenantID = createTenant(server, client)
		projID = createProject(server, client, tenantID)
		networkCreateSpec = &NetworkCreateSpec{Name: "network-1", PrivateIpCidr: "cidr1"}
	})

	AfterEach(func() {
		cleanTenants(client)
		server.Close()
	})

	Describe("CreateDeleteNetwork", func() {
		It("Network create and delete succeeds", func() {
			mockTask := createMockTask("CREATE_NETWORK", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.CreateNetwork(projID, networkCreateSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_NETWORK"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("DELETE_NETWORK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Networks.Delete("networkId")
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_NETWORK"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("GetNetwork", func() {
		It("Get returns network", func() {
			mockTask := createMockTask("CREATE_NETWORK", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Projects.CreateNetwork(projID, networkCreateSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_NETWORK"))
			Expect(task.State).Should(Equal("COMPLETED"))

			server.SetResponseJson(200, Network{Name: "network-1", PrivateIpCidr: "cidr1"})
			network, err := client.Networks.Get(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(network).ShouldNot(BeNil())

			var found bool
			if network.Name == "network-1" && network.ID == task.Entity.ID {
				found = true
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_NETWORK", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			_, err = client.Networks.Delete(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("UpdateNetwork", func() {
		It("update network's name", func() {
			mockTask := createMockTask("UPDATE_NETWORK", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			networkSpec := &NetworkUpdateSpec{NetworkName: "network-1"}
			task, err := client.Networks.UpdateNetwork("network-Id", networkSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("UPDATE_NETWORK"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("GetNetworkSubnets", func() {
		It("GetAll returns subnet", func() {
			mockTask := createMockTask("CREATE_SUBNET", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			subnetSpec := &SubnetCreateSpec{
				Name:          "subnet_name",
				Description:   "subnet description",
				PrivateIpCidr: "192.168.0.1/24",
			}

			task, err := client.Networks.CreateSubnet("network-id", subnetSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())

			subnetMock := Subnet{
				Name:          subnetSpec.Name,
				Description:   subnetSpec.Description,
				PrivateIpCidr: subnetSpec.PrivateIpCidr,
			}
			server.SetResponseJson(200, &Subnets{[]Subnet{subnetMock}})
			subnetList, err := client.Networks.GetSubnets("network-id", &SubnetGetOptions{})
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
