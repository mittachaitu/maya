package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	volume "github.com/openebs/maya/pkg/client/volume/cstor/v1alpha1"
)

// client is used to invoke cstor resize client call
type client struct {
	ip     string
	volume *apis.CASVolume
}

type clientBuilder struct {
	client *client
}

// ClientBuilder forms the populates the client structure
func ClientBuilder() *clientBuilder {
	return &clientBuilder{client: &client{}}
}

func (c *clientBuilder) WithIP(ip string) *clientBuilder {
	c.client.ip = ip
	return c
}

func (c *clientBuilder) WithVolume(v *apis.CASVolume) *clientBuilder {
	c.client.volume = v
	return c
}

// Build return a pointer to client
func (c *clientBuilder) Build() *client {
	return c.client
}

// Resize will trigger a grpc call to cstor-volume-mgmt side car to resize
// volume
func (c *client) Resize() (*apis.CASVolume, error) {
	_, err := volume.ResizeVolume(c.ip, c.volume.Name, c.volume.Spec.Capacity)
	if err != nil {
		return nil, err
	}

	// we are returning the same struct that we received as input.
	// This would be modified when server replies back with some property of
	// resize volume
	return c.volume, nil
}
