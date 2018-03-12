package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wearefair/k8-cross-cluster-controller/controller"
)

func runController(cmd *cobra.Command, args []string) {
	if err := controller.Coordinate(); err != nil {
		fmt.Println(err)
	}
}
