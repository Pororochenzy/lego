package m_comps

import (
	"github.com/Pororochenzy/lego/core"
	"github.com/Pororochenzy/lego/sys/proto"
)

type (
	IMComp_GateComp interface {
		core.IModuleComp
		ReceiveMsg(session core.IUserSession, msg proto.IMessage) (code core.ErrorCode, err string)
	}
)
