package main

import (
	"os"
	"runtime"

	"stash.appscode.dev/postgres/pkg"
	_ "stash.appscode.dev/stash/client/clientset/versioned/fake"

	"github.com/appscode/go/log"
	_ "k8s.io/client-go/kubernetes/fake"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"kmodules.xyz/client-go/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if err := pkg.NewRootCmd().Execute(); err != nil {
		log.Fatalln("error:", err)
	}
}
