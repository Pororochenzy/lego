package s_comps

import (
	"github.com/Pororochenzy/lego/core"
	"github.com/Pororochenzy/lego/sys/proto"
)

type (
	ISC_GateRouteComp interface {
		core.IServiceComp
		RegisterRoute(comId uint16, f func(s core.IUserSession, msg proto.IMessage) (code core.ErrorCode, err string)) (err error)
	}
)

func NewGateRouteComp() ISC_GateRouteComp {
	comp := new(SComp_GateRouteComp)
	return comp
}
