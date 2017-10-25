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

var _ = Describe("Datastores", func() {
	var (
		server        *mocks.Server
		client        *Client
		datastoreSpec *Datastore
	)

	BeforeEach(func() {
		server, client = testSetup()

		datastoreId := "1234"
		datastoreSpec = &Datastore{
			Kind:     "datastore",
			Type:     "LOCAL_VMFS",
			ID:       datastoreId,
			SelfLink: "https://192.0.0.2" + rootUrl + "/infrastructure/datastores/" + datastoreId,
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Get", func() {
		It("Get a single datastore successfully", func() {

			server.SetResponseJson(200, datastoreSpec)

			datastore, err := client.Datastores.Get(datastoreSpec.ID)
			GinkgoT().Log(err)

			Expect(err).Should(BeNil())
			Expect(datastore).ShouldNot(BeNil())
			Expect(datastore.Kind).Should(Equal(datastoreSpec.Kind))
			Expect(datastore.Type).Should(Equal(datastoreSpec.Type))
			Expect(datastore.ID).Should(Equal(datastoreSpec.ID))

		})
	})
	Describe("GetAll", func() {
		It("Get all datastores successfully", func() {
			datastoresExpected := Datastores{
				Items: []Datastore{*datastoreSpec},
			}

			server.SetResponseJson(200, datastoresExpected)

			datastores, err := client.Datastores.GetAll()
			GinkgoT().Log(err)

			Expect(err).Should(BeNil())
			Expect(datastores).ShouldNot(BeNil())
			Expect(datastores.Items[0].Kind).Should(Equal(datastoreSpec.Kind))
			Expect(datastores.Items[0].Type).Should(Equal(datastoreSpec.Type))
			Expect(datastores.Items[0].ID).Should(Equal(datastoreSpec.ID))
		})
	})
})
