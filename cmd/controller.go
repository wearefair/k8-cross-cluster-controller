package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wearefair/k8-cross-cluster-controller/controller"
	"github.com/wearefair/k8-cross-cluster-controller/utils"
)

func runController(cmd *cobra.Command, args []string) {
	conf, err := utils.PathHelper(kubeconfig)
	if err != nil {
		return err
	}
	if err := controller.Coordinate(conf); err != nil {
		return err
	}
	return nil
}
