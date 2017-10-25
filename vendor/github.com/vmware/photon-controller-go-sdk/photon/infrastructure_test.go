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

var _ = Describe("Infra", func() {
	var (
		server *mocks.Server
		client *Client
	)

	BeforeEach(func() {
		if isIntegrationTest() {
			Skip("Skipping deployment test on integration mode. Need undeployed environment")
		}
		server, client = testSetup()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("SyncHostsConfig", func() {
		It("Sync Hosts Config succeeds", func() {
			mockTask := createMockTask("SYNC_HOSTS_CONFIG", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Infra.SyncHostsConfig()
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("SYNC_HOSTS_CONFIG"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("SetImageDatastores", func() {
		It("Succeeds", func() {
			mockTask := createMockTask("UPDATE_IMAGE_DATASTORES", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			imageDatastores := &ImageDatastores{
				[]string{"imageDatastore1", "imageDatastore2"},
			}
			createdTask, err := client.Infra.SetImageDatastores(imageDatastores)
			createdTask, err = client.Tasks.Wait(createdTask.ID)

			Expect(err).Should(BeNil())
			Expect(createdTask.Operation).Should(Equal("UPDATE_IMAGE_DATASTORES"))
			Expect(createdTask.State).Should(Equal("COMPLETED"))
		})

		It("Fails", func() {
			mockApiError := createMockApiError("INVALID_IMAGE_DATASTORES", "Not a super set", 400)
			server.SetResponseJson(400, mockApiError)

			imageDatastores := &ImageDatastores{
				[]string{"imageDatastore1", "imageDatastore2"},
			}
			createdTask, err := client.Infra.SetImageDatastores(imageDatastores)

			Expect(err).Should(Equal(*mockApiError))
			Expect(createdTask).Should(BeNil())
		})
	})
})
