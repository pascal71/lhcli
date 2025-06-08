package cmd

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pascal71/lhcli/pkg/formatter"
	"github.com/spf13/cobra"
)

var pvCmd = &cobra.Command{
	Use:   "pv",
	Short: "Interact with Kubernetes PersistentVolumes",
	Long:  `Operations related to Kubernetes PersistentVolumes used by Longhorn.`,
}

var pvMapCmd = &cobra.Command{
	Use:   "map [pv-name]",
	Short: "Map Kubernetes PVs to Longhorn volumes",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runPVMap,
}

func init() {
	rootCmd.AddCommand(pvCmd)
	pvCmd.AddCommand(pvMapCmd)
}

func runPVMap(cmd *cobra.Command, args []string) error {
	kubeClient, err := getKubeClient()
	if err != nil {
		return err
	}

	pvs, err := kubeClient.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list persistent volumes: %w", err)
	}

	filter := ""
	if len(args) == 1 {
		filter = args[0]
	}

	type mapping struct {
		PV        string `json:"pv"`
		Volume    string `json:"volume"`
		PVC       string `json:"pvc,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}

	var mappings []mapping
	for _, pv := range pvs.Items {
		if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != "driver.longhorn.io" {
			continue
		}
		if filter != "" && pv.Name != filter {
			continue
		}
		m := mapping{
			PV:     pv.Name,
			Volume: pv.Spec.CSI.VolumeHandle,
		}
		if pv.Spec.ClaimRef != nil {
			m.PVC = pv.Spec.ClaimRef.Name
			m.Namespace = pv.Spec.ClaimRef.Namespace
		}
		mappings = append(mappings, m)
	}

	switch output {
	case "json":
		return formatter.NewJSONFormatter(true).Format(mappings)
	case "yaml":
		return formatter.NewYAMLFormatter().Format(mappings)
	default:
		headers := []string{"PV", "VOLUME", "PVC", "NAMESPACE"}
		table := formatter.NewTableFormatter(headers)
		for _, m := range mappings {
			table.AddRow([]string{m.PV, m.Volume, m.PVC, m.Namespace})
		}
		return table.Format(nil)
	}
}
