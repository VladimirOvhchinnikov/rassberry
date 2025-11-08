package main

import (
	"example.com/ffp/kernels/examples/site"
	rt "example.com/ffp/platform/runtime"
)

func init() {
	RegisterDomainFactory("site", func(id string) rt.KernelModule {
		return site.NewDomain(id)
	})
}
