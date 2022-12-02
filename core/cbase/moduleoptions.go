package cbase

import "github.com/Pororochenzy/lego/utils/mapstructure"

type ModuleOptions struct {
}

func (this *ModuleOptions) LoadConfig(settings map[string]interface{}) (err error) {
	if settings != nil {
		mapstructure.Decode(settings, &this)
	}
	return
}
