package terraform

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"io"
	"log"
	"os"
)

func Copy(srcpath, dstpath string) (err error) {
	r, err := os.Open(srcpath)
	if err != nil {
		return err
	}
	defer func(r *os.File) {
		closeErr := r.Close()
		if closeErr != nil {
			log.Println("failed to close file handle. details:", err)
		}
	}(r) // ok to ignore error: file was opened read-only.

	w, err := os.Create(dstpath)
	if err != nil {
		return err
	}

	defer func() {
		c := w.Close()
		// Report the error from Close, if any.
		// But do so only if there isn't already
		// an outgoing error.
		if c != nil && err == nil {
			err = c
		}
	}()

	_, err = io.Copy(w, r)
	return err
}

func CreateTfResource(resourceType string, localName string, values map[string]cty.Value) (*hclwrite.File, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()

	// resource
	{
		ep := rootBody.AppendNewBlock("resource", []string{resourceType, localName})
		epBody := ep.Body()
		for key, value := range values {
			epBody.SetAttributeValue(key, value)
		}
	}

	rootBody.AppendNewline()

	return f, nil
}

func CreateTfSystemConf(providers []Provider) (*hclwrite.File, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()

	// resource
	{
		terraform := rootBody.AppendNewBlock("terraform", nil)
		reqProviders := terraform.Body().AppendNewBlock("required_providers", nil).Body()

		for _, provider := range providers {
			reqProviders.SetAttributeValue(provider.Name, cty.MapVal(map[string]cty.Value{
				"source":  cty.StringVal(provider.Source),
				"version": cty.StringVal(provider.Version),
			}))
		}
	}

	rootBody.AppendNewline()

	return f, nil
}
