package cbase

import (
	"github.com/Pororochenzy/lego/core"
	"github.com/Pororochenzy/lego/utils/mapstructure"
)

type ServiceCompOptions struct {
	core.ICompOptions
}

func (this *ServiceCompOptions) LoadConfig(settings map[string]interface{}) (err error) {
	if settings != nil {
		mapstructure.Decode(settings, &this)
	}
	return
}
