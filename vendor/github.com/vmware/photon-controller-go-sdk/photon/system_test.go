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
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/photon-controller-go-sdk/photon/internal/mocks"
)

var _ = Describe("System", func() {
	var (
		server *mocks.Server
		client *Client
	)

	BeforeEach(func() {
		if isIntegrationTest() {
			Skip("Skipping system test on integration mode.")
		}
		server, client = testSetup()
	})

	AfterEach(func() {
		server.Close()
	})

	// Tests system status
	Describe("GetSystemStatus", func() {
		It("GetStatus200", func() {
			expectedStruct := Status{"READY", []Component{{"PHOTON_CONTROLLER", "", "READY"}}}
			server.SetResponseJson(200, expectedStruct)

			status, err := client.System.GetSystemStatus()
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(status.Status).Should(Equal(expectedStruct.Status))
			Expect(status.Components).Should(HaveLen(1))
		})
	})

	// Tests system info
	Describe("GetSystemInfo", func() {
		It("GetInfoFromDeployment", func() {
			mockDeployment := SystemInfo{
				ImageDatastores:         []string{randomString(10, "go-sdk-deployment-")},
				UseImageDatastoreForVms: true,
				Auth:                 &AuthInfo{},
				NetworkConfiguration: &NetworkConfiguration{Enabled: false},
			}
			server.SetResponseJson(200, mockDeployment)
			deployment, err := client.System.GetSystemInfo()

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(deployment).ShouldNot(BeNil())
			Expect(deployment.ImageDatastores).Should(Equal(mockDeployment.ImageDatastores))
			Expect(deployment.UseImageDatastoreForVms).Should(Equal(mockDeployment.UseImageDatastoreForVms))
		})
	})

	// Tests pause and resume system
	Describe("PauseSystemAndPauseBackgroundTasks", func() {
		It("Pause System and Resume System succeeds", func() {
			mockTask := createMockTask("PAUSE_SYSTEM", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.System.PauseSystem()
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("PAUSE_SYSTEM"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("RESUME_SYSTEM", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err = client.System.ResumeSystem()
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("RESUME_SYSTEM"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Pause Background Tasks and Resume System succeeds", func() {
			mockTask := createMockTask("PAUSE_BACKGROUND_TASKS", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.System.PauseBackgroundTasks()
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("PAUSE_BACKGROUND_TASKS"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("RESUME_SYSTEM", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err = client.System.ResumeSystem()
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("RESUME_SYSTEM"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	// Tests update to system security group
	Describe("SetSecurityGroups", func() {
		It("sets security groups for the system", func() {
			mockTask := createMockTask("SET_SECURITY_GROUPS", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			// Set security groups for the system
			expected := &SystemInfo{
				Auth: &AuthInfo{
					SecurityGroups: []string{
						randomString(10),
						randomString(10),
					},
				},
			}

			payload := SecurityGroupsSpec{
				Items: expected.Auth.SecurityGroups,
			}
			updateTask, err := client.System.SetSecurityGroups(&payload)
			updateTask, err = client.Tasks.Wait(updateTask.ID)
			Expect(err).Should(BeNil())

			// Get the security groups for the system
			server.SetResponseJson(200, expected)
			deployment, err := client.System.GetSystemInfo()
			Expect(err).Should(BeNil())
			Expect(deployment.Auth.SecurityGroups).To(ContainElement(payload.Items[0]))
			Expect(deployment.Auth.SecurityGroups).To(ContainElement(payload.Items[1]))
		})
	})

	// Tests system size
	Describe("SystemSize", func() {
		It("GetSystemSize", func() {
			mockSize := &SystemUsage{
				NumberDatastores: 1,
				NumberHosts:      1,
				NumberProjects:   0,
				NumberServices:   1,
				NumberTenants:    2,
			}
			server.SetResponseJson(200, mockSize)

			systemSize, err := client.System.GetSystemSize()
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(systemSize).ShouldNot(BeNil())
			Expect(systemSize.NumberDatastores).Should(Equal(1))
			Expect(systemSize.NumberHosts).Should(Equal(1))
			Expect(systemSize.NumberProjects).Should(Equal(0))
			Expect(systemSize.NumberServices).Should(Equal(1))
			Expect(systemSize.NumberTenants).Should(Equal(2))
		})
	})

	Describe("GetAuthInfo", func() {
		It("returns auth info", func() {
			expected := createMockAuthInfo(nil)
			server.SetResponseJson(200, expected)

			info, err := client.System.GetAuthInfo()
			fmt.Fprintf(GinkgoWriter, "Got auth info: %+v\n", info)
			Expect(err).Should(BeNil())
			Expect(info).ShouldNot(BeNil())
			Expect(info).Should(BeEquivalentTo(expected))
		})
	})

	Describe("EnableAndDisableServiceType", func() {
		It("Enable And Disable Service Type", func() {
			serviceType := "SWARM"
			serviceImageId := "testImageId"
			serviceConfigSpec := &ServiceConfigurationSpec{
				Type:    serviceType,
				ImageID: serviceImageId,
			}

			mockTask := createMockTask("CONFIGURE_SERVICE", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			enableTask, err := client.System.EnableServiceType(serviceConfigSpec)
			enableTask, err = client.Tasks.Wait(enableTask.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(enableTask).ShouldNot(BeNil())
			Expect(enableTask.Operation).Should(Equal("CONFIGURE_SERVICE"))
			Expect(enableTask.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("DELETE_SERVICE_CONFIGURATION", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			disableTask, err := client.System.DisableServiceType(serviceConfigSpec)
			disableTask, err = client.Tasks.Wait(disableTask.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(disableTask).ShouldNot(BeNil())
			Expect(disableTask.Operation).Should(Equal("DELETE_SERVICE_CONFIGURATION"))
			Expect(disableTask.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("ConfigureNsx", func() {
		It("Configure NSX", func() {
			nsxAddress := "nsxAddress"
			nsxUsername := "nsxUsername"
			nsxPassword := "nsxPassword"

			nsxConfigSpec := &NsxConfigurationSpec{
				NsxAddress:  nsxAddress,
				NsxUsername: nsxUsername,
				NsxPassword: nsxPassword,
			}

			mockTask := createMockTask("CONFIGURE_NSX", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			enableTask, err := client.System.ConfigureNsx(nsxConfigSpec)
			enableTask, err = client.Tasks.Wait(enableTask.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(enableTask).ShouldNot(BeNil())
			Expect(enableTask.Operation).Should(Equal("CONFIGURE_NSX"))
			Expect(enableTask.State).Should(Equal("COMPLETED"))
		})
	})
})
