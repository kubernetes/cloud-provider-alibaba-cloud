package crd

import (
	"context"
	"fmt"
	"time"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextcli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	checkCRDInterval = 2 * time.Second
	crdReadyTimeout  = 3 * time.Minute
)

var (
	defCategories = []string{"all", "kooper"}
	log           = klogr.New().WithName("crd")
)

// Scope is the scope of a CRD.
type Scope = apiextv1.ResourceScope

const (
	// ClusterScoped represents a type of a cluster scoped CRD.
	ClusterScoped = apiextv1.ClusterScoped
	// NamespaceScoped represents a type of a namespaced scoped CRD.
	NamespaceScoped = apiextv1.NamespaceScoped
)

// Conf is the configuration required to create a CRD
type Conf struct {
	// Kind is the kind of the CRD.
	Kind string
	// NamePlural is the plural name of the CRD (in most cases the plural of Kind).
	NamePlural string
	// ShortNames are short names of the CRD.  It must be all lowercase.
	ShortNames []string
	// Group is the group of the CRD.
	Group string
	// Version is the version of the CRD.
	Version string
	// Scope is the scode of the CRD (cluster scoped or namespace scoped).
	Scope Scope
	// Categories is a way of grouping multiple resources (example `kubectl get all`),
	// Kooper adds the CRD to `all` and `kooper` categories(apart from the described in Caregories).
	Categories []string
	// EnableStatus will enable the Status subresource on the CRD. This is feature
	// entered in v1.10 with the CRD subresources.
	// By default is disabled.
	EnableStatusSubresource bool
	// EnableScaleSubresource by default will be nil and means disabled, if
	// the object is present it will set this scale configuration to the subresource.
	EnableScaleSubresource *apiextv1.CustomResourceSubresourceScale
}

func (c *Conf) getName() string {
	return fmt.Sprintf("%s.%s", c.NamePlural, c.Group)
}

// Interface is the CRD client that knows how to interact with k8s to manage them.
type Interface interface {
	// EnsureCreated will ensure the the CRD is present, this also means that
	// apart from creating the CRD if is not present it will wait until is
	// ready, this is a blocking operation and will return an error if timesout
	// waiting.
	EnsurePresent(conf Conf) error
	// WaitToBePresent will wait until the CRD is present, it will check if
	// is present at regular intervals until it timesout, in case of timeout
	// will return an error.
	WaitToBePresent(name string, timeout time.Duration) error
	// Delete will delete the CRD.
	Delete(name string) error
}

// ECS is the CRD client implementation using API calls to kubernetes.
type Client struct {
	client apiextcli.Interface
}

// NewClient returns a new CRD client.
func NewClient(client apiextcli.Interface) *Client {
	return NewCustomClient(client)
}

// NewCustomClient returns a new CRD client letting you set all the required parameters
func NewCustomClient(
	client apiextcli.Interface,
) *Client {
	return &Client{
		client: client,
	}
}

// EnsurePresent satisfies crd.Interface.
func (c *Client) EnsurePresent(conf Conf) error {
	err := c.validateCRD()
	if err != nil {
		return fmt.Errorf("validate crd: %s", err.Error())
	}

	// Get the generated name of the CRD.
	crdName := conf.getName()

	// Create subresources
	subres := c.createSubresources(conf)
	xPreserveUnknownFields := true
	openAPIV3Schema := &apiextv1.JSONSchemaProps{
		XPreserveUnknownFields: &xPreserveUnknownFields,
	}
	schema := &apiextv1.CustomResourceValidation{OpenAPIV3Schema: openAPIV3Schema}
	crd := &apiextv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
		Spec: apiextv1.CustomResourceDefinitionSpec{
			Group: conf.Group,
			Versions: []apiextv1.CustomResourceDefinitionVersion{{Name: conf.Version, Served: true, Storage: true, Subresources: subres, Schema: schema,
				AdditionalPrinterColumns: []apiextv1.CustomResourceColumnDefinition{
					{
						Name:     "ALBID",
						Type:     "string",
						JSONPath: ".status.loadBalancer.id",
					},
					{
						Name:     "DNSNAME",
						Type:     "string",
						JSONPath: ".status.loadBalancer.dnsname",
					},
					{
						Name:     "PORT&PROTOCOL",
						Type:     "string",
						JSONPath: ".status.loadBalancer.listeners[*].portAndProtocol",
					},
					{
						Name:     "CERTID",
						Type:     "string",
						JSONPath: ".status.loadBalancer.listeners[*].certificates[*].certificateId",
					},
					{
						Name:     "AGE",
						Type:     "date",
						JSONPath: ".metadata.creationTimestamp",
					},
				}}},
			Scope: conf.Scope,
			Names: apiextv1.CustomResourceDefinitionNames{
				Plural:     conf.NamePlural,
				Kind:       conf.Kind,
				ShortNames: conf.ShortNames,
				Categories: c.addDefaultCaregories(conf.Categories),
			},
		},
	}
	original, err := c.client.ApiextensionsV1().
		CustomResourceDefinitions().Get(context.TODO(), crd.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = c.client.ApiextensionsV1().
				CustomResourceDefinitions().
				Create(context.TODO(), crd, metav1.CreateOptions{})
			if err != nil {
				klog.Errorf("error creating crd %s: %s", crdName, err)
				return err
			}
			klog.Info(fmt.Sprintf("crd %s created, waiting to be ready...", crdName))
			return c.WaitToBePresent(crdName, crdReadyTimeout)
		}
		return err
	}
	if !apiequality.Semantic.DeepEqual(original.Spec.Versions, &crd.Spec.Versions) {
		original.Spec.Versions = crd.Spec.Versions
		_, err = c.client.ApiextensionsV1().
			CustomResourceDefinitions().Update(context.TODO(), original, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("error Update crd %s: %s", crdName, err)
			return err
		}
	}
	return nil
}

func (c *Client) createSubresources(conf Conf) *apiextv1.CustomResourceSubresources {
	if !conf.EnableStatusSubresource &&
		conf.EnableScaleSubresource == nil {
		return nil
	}

	sr := &apiextv1.CustomResourceSubresources{}

	if conf.EnableStatusSubresource {
		sr.Status = &apiextv1.CustomResourceSubresourceStatus{}
	}

	if conf.EnableScaleSubresource != nil {
		sr.Scale = conf.EnableScaleSubresource
	}
	return sr
}

// WaitToBePresent satisfies crd.Interface.
func (c *Client) WaitToBePresent(name string, timeout time.Duration) error {
	err := c.validateCRD()
	if err != nil {
		return fmt.Errorf("wait validate crd: %s", err.Error())
	}

	tick := time.NewTicker(checkCRDInterval)
	for {
		select {
		case <-tick.C:
			_, err := c.client.
				ApiextensionsV1().
				CustomResourceDefinitions().
				Get(
					context.TODO(), name, metav1.GetOptions{},
				)
			// Is present, finish.
			if err == nil {
				return nil
			}
		case <-time.After(timeout):
			return fmt.Errorf("timeout waiting for CRD")
		}
	}
}

// Delete satisfies crd.Interface.
func (c *Client) Delete(name string) error {
	err := c.validateCRD()
	if err != nil {
		return fmt.Errorf("validate crd: %s", err.Error())
	}

	return c.client.
		ApiextensionsV1().
		CustomResourceDefinitions().
		Delete(
			context.TODO(), name, metav1.DeleteOptions{},
		)
}

// validateCRD returns nil if cluster is ok to be used for CRDs, otherwise error.
func (c *Client) validateCRD() error {
	// Check cluster version
	serverVersion, err := c.client.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("get server version: %s", err.Error())
	}

	runningVersion, err := version.ParseGeneric(serverVersion.String())
	if err != nil {
		return fmt.Errorf("unexpected error parsing running Kubernetes version, %s", err.Error())
	}

	leastVersion, _ := version.ParseGeneric("v1.19.0")
	if !runningVersion.AtLeast(leastVersion) {
		log.Info("kubernetes version should great then v1.19.0 to use crd", "server version", serverVersion.GitVersion)
	}

	return nil
}

// addAllCaregory adds the `all` category if isn't present
func (c *Client) addDefaultCaregories(categories []string) []string {
	currentCats := make(map[string]bool)
	for _, ca := range categories {
		currentCats[ca] = true
	}

	// Add default categories if required.
	for _, ca := range defCategories {
		if _, ok := currentCats[ca]; !ok {
			categories = append(categories, ca)
		}
	}

	return categories
}
