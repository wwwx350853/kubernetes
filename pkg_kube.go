package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/pflag"
	//"k8s.io/kubernetes/pkg/api"
	//"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util"
	"k8s.io/kubernetes/pkg/util/flag"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/version/verflag"
	"k8s.io/kubernetes/plugin/pkg/testclient/etcdio"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	e := &etcdio.EtcdIo{}
	e.AddFlags(pflag.CommandLine)

	flag.InitFlags()
	util.InitLogs()
	defer util.FlushLogs()

	verflag.PrintAndExitIfRequested()

	if err := e.Initial(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	e.Run(wait.NeverStop)

}
