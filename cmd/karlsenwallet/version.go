package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/karlsen-network/karlsend/v2/version"
)

func showVersion() {
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	fmt.Println(appName, "version", version.Version())
}
