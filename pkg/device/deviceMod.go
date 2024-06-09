package device

import "github.com/alaingilbert/ogame/pkg/httpclient"

func (d *Builder) SetClient(client *httpclient.Client) *Builder {
	d.client = client
	return d
}

func (d *Builder) GetNewFingerprint() (*JsFingerprint, error) {
	return d.newFingerprint()
}
