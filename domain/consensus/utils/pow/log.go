package pow

import (
	"github.com/karlsen-network/karlsend/infrastructure/logger"
	"github.com/karlsen-network/karlsend/util/panics"
)

// var log = logger.RegisterSubSystem("POW")
var backendLog = logger.NewBackend()
var log = backendLog.Logger("POW")
var spawn = panics.GoroutineWrapperFunc(log)
