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
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/photon-controller-go-sdk/photon/internal/mocks"
)

var _ = Describe("Tenant", func() {
	var (
		server   *mocks.Server
		client   *Client
		tenantID string
	)

	BeforeEach(func() {
		server, client = testSetup()
	})

	AfterEach(func() {
		cleanTenants(client)
		server.Close()
	})

	Describe("CreateAndDeleteTenant", func() {
		It("Tenant create and delete succeeds", func() {
			mockTask := createMockTask("CREATE_TENANT", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			tenantSpec := &TenantCreateSpec{Name: randomString(10, "go-sdk-tenant-")}
			task, err := client.Tenants.Create(tenantSpec)
			task, err = client.Tasks.Wait(task.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_TENANT"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTask = createMockTask("DELETE_TENANT", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Tenants.Delete(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_TENANT"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Tenant create fails", func() {
			tenantSpec := &TenantCreateSpec{}
			task, err := client.Tenants.Create(tenantSpec)

			server.SetResponseJson(200, createMockTenantsPage())

			Expect(err).ShouldNot(BeNil())
			Expect(task).Should(BeNil())
		})

		It("creates and deletes a tenant with security groups specified", func() {
			// Create a tenant with security groups.
			mockTask := createMockTask("CREATE_TENANT", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			securityGroups := []string{randomString(10), randomString(10)}
			expected := &Tenant{
				SecurityGroups: []SecurityGroup{
					{securityGroups[0], false},
					{securityGroups[1], false},
				},
			}
			tenantSpec := &TenantCreateSpec{
				Name:           randomString(10, "go-sdk-tenant-"),
				SecurityGroups: securityGroups,
			}
			task, err := client.Tenants.Create(tenantSpec)
			Expect(task).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
			task, err = client.Tasks.Wait(task.ID)
			Expect(task).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_TENANT"))
			Expect(task.State).Should(Equal("COMPLETED"))
			tenantId := task.Entity.ID

			// Verify that the tenant has the security groups set properly.
			server.SetResponseJson(200, expected)
			tenant, err := client.Tenants.Get(tenantId)
			Expect(err).Should(BeNil())
			Expect(tenant.SecurityGroups).Should(Equal(expected.SecurityGroups))

			// Delete the tenant.
			mockTask = createMockTask("DELETE_TENANT", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Tenants.Delete(tenantId)
			Expect(task).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_TENANT"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("GetTenant", func() {
		It("Get returns a tenant ID", func() {
			mockTask := createMockTask("CREATE_TENANT", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			tenantName := randomString(10, "go-sdk-tenant-")
			tenantSpec := &TenantCreateSpec{Name: tenantName}
			task, err := client.Tenants.Create(tenantSpec)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_TENANT"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockTenantsPage := createMockTenantsPage(Tenant{Name: tenantName})
			server.SetResponseJson(200, mockTenantsPage)
			tenants, err := client.Tenants.GetAll()

			mockTenantPage := createMockTenantPage()
			server.SetResponseJson(200, mockTenantPage)
			mockTenant, err := client.Tenants.Get("TestTenant")
			Expect(mockTenant.Name).Should(Equal("TestTenant"))
			Expect(mockTenant.ID).Should(Equal("12345"))

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(tenants).ShouldNot(BeNil())

			var found bool
			for _, tenant := range tenants.Items {
				if tenant.Name == tenantName && tenant.ID == task.Entity.ID {
					found = true
					break
				}
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_TENANT", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			_, err = client.Tenants.Delete(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
		})
	})

	Describe("GetTenantTasks", func() {
		var (
			option string
		)

		Context("no extra options for GetTask", func() {
			BeforeEach(func() {
				option = ""
			})

			It("GetTasks returns a completed task", func() {
				mockTask := createMockTask("CREATE_TENANT", "COMPLETED")
				mockTask.Entity.ID = "mock-task-id"
				server.SetResponseJson(200, mockTask)
				tenantSpec := &TenantCreateSpec{Name: randomString(10, "go-sdk-tenant-")}
				task, err := client.Tenants.Create(tenantSpec)

				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
				Expect(task).ShouldNot(BeNil())
				Expect(task.Operation).Should(Equal("CREATE_TENANT"))
				Expect(task.State).Should(Equal("COMPLETED"))

				mockTasksPage := createMockTasksPage(*mockTask)
				server.SetResponseJson(200, mockTasksPage)
				taskList, err := client.Tenants.GetTasks(task.Entity.ID, &TaskGetOptions{State: option})
				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
				Expect(taskList).ShouldNot(BeNil())
				Expect(taskList.Items).Should(ContainElement(*task))

				mockTask = createMockTask("DELETE_TENANT", "COMPLETED")
				server.SetResponseJson(200, mockTask)
				_, err = client.Tenants.Delete(task.Entity.ID)

				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
			})
		})

		Context("Searching COMPLETED state for GetTask", func() {
			BeforeEach(func() {
				option = "COMPLETED"
			})

			It("GetTasks returns a completed task", func() {
				mockTask := createMockTask("CREATE_TENANT", "COMPLETED")
				mockTask.Entity.ID = "mock-task-id"
				server.SetResponseJson(200, mockTask)
				tenantSpec := &TenantCreateSpec{Name: randomString(10, "go-sdk-tenant-")}
				task, err := client.Tenants.Create(tenantSpec)

				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
				Expect(task).ShouldNot(BeNil())
				Expect(task.Operation).Should(Equal("CREATE_TENANT"))
				Expect(task.State).Should(Equal("COMPLETED"))

				mockTasksPage := createMockTasksPage(*mockTask)
				server.SetResponseJson(200, mockTasksPage)
				taskList, err := client.Tenants.GetTasks(task.Entity.ID, &TaskGetOptions{State: option})
				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
				Expect(taskList).ShouldNot(BeNil())
				Expect(taskList.Items).Should(ContainElement(*task))

				mockTask = createMockTask("DELETE_TENANT", "COMPLETED")
				server.SetResponseJson(200, mockTask)
				_, err = client.Tenants.Delete(task.Entity.ID)

				GinkgoT().Log(err)
				Expect(err).Should(BeNil())
			})
		})

	})

	Describe("TenantQuota", func() {

		It("Get Tenant Quota succeeds", func() {
			mockQuota := createMockQuota()

			// Create Tenant
			tenantID = createTenant(server, client)

			// Get current Quota
			server.SetResponseJson(200, mockQuota)
			quota, err := client.Tenants.GetQuota(tenantID)

			GinkgoT().Log(err)
			eq := reflect.DeepEqual(quota.QuotaLineItems, mockQuota.QuotaLineItems)
			Expect(eq).Should(Equal(true))
		})

		It("Set Tenant Quota succeeds", func() {
			mockQuotaSpec := &QuotaSpec{
				"vmCpu":        {Unit: "COUNT", Limit: 10, Usage: 0},
				"vmMemory":     {Unit: "GB", Limit: 18, Usage: 0},
				"diskCapacity": {Unit: "GB", Limit: 100, Usage: 0},
			}

			// Create Tenant
			tenantID = createTenant(server, client)

			mockTask := createMockTask("SET_QUOTA", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Tenants.SetQuota(tenantID, mockQuotaSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("SET_QUOTA"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Update Tenant Quota succeeds", func() {
			mockQuotaSpec := &QuotaSpec{
				"vmCpu":        {Unit: "COUNT", Limit: 30, Usage: 0},
				"vmMemory":     {Unit: "GB", Limit: 40, Usage: 0},
				"diskCapacity": {Unit: "GB", Limit: 150, Usage: 0},
			}

			// Create Tenant
			tenantID = createTenant(server, client)

			mockTask := createMockTask("UPDATE_QUOTA", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Tenants.UpdateQuota(tenantID, mockQuotaSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("UPDATE_QUOTA"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})

		It("Exclude Tenant Quota Items succeeds", func() {
			mockQuotaSpec := &QuotaSpec{
				"vmCpu2":    {Unit: "COUNT", Limit: 10, Usage: 0},
				"vmMemory3": {Unit: "GB", Limit: 18, Usage: 0},
			}

			// Create Tenant
			tenantID = createTenant(server, client)

			mockTask := createMockTask("DELETE_QUOTA", "COMPLETED")
			server.SetResponseJson(200, mockTask)

			task, err := client.Tenants.ExcludeQuota(tenantID, mockQuotaSpec)
			task, err = client.Tasks.Wait(task.ID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_QUOTA"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})
})

var _ = Describe("Project", func() {
	var (
		server   *mocks.Server
		client   *Client
		tenantID string
	)

	BeforeEach(func() {
		server, client = testSetup()
		tenantID = createTenant(server, client)
	})

	AfterEach(func() {
		cleanTenants(client)
		server.Close()
	})

	Describe("CreateGetAndDeleteProject", func() {
		It("Project create and delete succeeds", func() {
			mockTask := createMockTask("CREATE_PROJECT", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			projSpec := &ProjectCreateSpec{
				Name: randomString(10, "go-sdk-project-"),
			}
			task, err := client.Tenants.CreateProject(tenantID, projSpec)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_PROJECT"))
			Expect(task.State).Should(Equal("COMPLETED"))

			mockProjects := createMockProjectsPage(ProjectCompact{Name: projSpec.Name})
			server.SetResponseJson(200, mockProjects)
			projList, err := client.Tenants.GetProjects(tenantID, &ProjectGetOptions{})
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(projList).ShouldNot(BeNil())

			var found bool
			for _, proj := range projList.Items {
				if proj.Name == projSpec.Name && proj.ID == task.Entity.ID {
					found = true
					break
				}
			}
			Expect(found).Should(BeTrue())

			mockTask = createMockTask("DELETE_PROJECT", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Projects.Delete(task.Entity.ID)

			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_PROJECT"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})

	Describe("SecurityGroups", func() {
		It("sets security groups for a tenant", func() {
			// Create a tenant
			mockTask := createMockTask("CREATE_TENANT", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			tenantSpec := &TenantCreateSpec{Name: randomString(10, "go-sdk-tenant-")}
			task, err := client.Tenants.Create(tenantSpec)
			task, err = client.Tasks.Wait(task.ID)
			Expect(err).Should(BeNil())
			Expect(task).ShouldNot(BeNil())
			Expect(task.Operation).Should(Equal("CREATE_TENANT"))
			Expect(task.State).Should(Equal("COMPLETED"))

			// Set security groups for the tenant
			expected := &Tenant{
				SecurityGroups: []SecurityGroup{
					{randomString(10), false},
					{randomString(10), false},
				},
			}
			mockTask = createMockTask("SET_TENANT_SECURITY_GROUPS", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			securityGroups := &SecurityGroupsSpec{
				[]string{expected.SecurityGroups[0].Name, expected.SecurityGroups[1].Name},
			}
			createTask, err := client.Tenants.SetSecurityGroups(task.Entity.ID, securityGroups)
			createTask, err = client.Tasks.Wait(createTask.ID)
			Expect(err).Should(BeNil())

			// Get the security groups for the tenant
			server.SetResponseJson(200, expected)
			tenant, err := client.Tenants.Get(task.Entity.ID)
			Expect(err).Should(BeNil())
			fmt.Fprintf(GinkgoWriter, "Got tenant: %+v", tenant)
			Expect(expected.SecurityGroups).Should(Equal(tenant.SecurityGroups))

			// Delete the tenant
			mockTask = createMockTask("DELETE_TENANT", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			task, err = client.Tenants.Delete(task.Entity.ID)
			task, err = client.Tasks.Wait(task.ID)
			Expect(err).Should(BeNil())
			Expect(task.Operation).Should(Equal("DELETE_TENANT"))
			Expect(task.State).Should(Equal("COMPLETED"))
		})
	})
})

var _ = Describe("IAM", func() {
	var (
		server   *mocks.Server
		client   *Client
		tenantID string
	)

	BeforeEach(func() {
		server, client = testSetup()
		tenantID = createTenant(server, client)
	})

	AfterEach(func() {
		cleanTenants(client)
		server.Close()
	})

	Describe("ManageTenantIamPolicy", func() {
		It("Set IAM Policy succeeds", func() {
			mockTask := createMockTask("SET_IAM_POLICY", "COMPLETED")
			server.SetResponseJson(200, mockTask)
			var policy []*RoleBinding
			policy = []*RoleBinding{{Role: "owner", Subjects: []string{"joe@photon.local"}}}
			task, err := client.Tenants.SetIam(tenantID, policy)
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
			task, err := client.Tenants.ModifyIam(tenantID, delta)
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
			response, err := client.Tenants.GetIam(tenantID)
			GinkgoT().Log(err)
			Expect(err).Should(BeNil())
			Expect(response[0].Subjects).Should(Equal(policy[0].Subjects))
			Expect(response[0].Role).Should(Equal(policy[0].Role))
		})
	})
})
