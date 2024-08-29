package connmanager

import (
	"github.com/karlsen-network/karlsend/v2/infrastructure/logger"
	"github.com/karlsen-network/karlsend/v2/util/panics"
)

var log = logger.RegisterSubSystem("CMGR")
var spawn = panics.GoroutineWrapperFunc(log)
