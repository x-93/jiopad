package flowcontext

import (
	"github.com/karlsen-network/karlsend/v2/domain"
)

// Domain returns the Domain object associated to the flow context.
func (f *FlowContext) Domain() domain.Domain {
	return f.domain
}
