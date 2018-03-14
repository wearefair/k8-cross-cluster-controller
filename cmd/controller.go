package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wearefair/k8-cross-cluster-controller/controller"
	"github.com/wearefair/k8-cross-cluster-controller/utils"
)

func runController(cmd *cobra.Command, args []string) {
	conf, err := utils.PathHelper(kubeconfig)
	if err != nil {
		panic(err)
	}
	internalConf, err := controller.SetupInternalConfig()
	if err != nil {
		panic(err)
	}
	remoteConf, err := controller.SetupRemoteConfig(conf)
	if err != nil {
		panic(err)
	}
	if err := controller.Coordinate(internalConf, remoteConf); err != nil {
		panic(err)
	}
}
