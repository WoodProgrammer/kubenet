package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	namespace string
	podName   string
	tcpFilter string
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-kubenet",
	Short: "A kubectl plugin to analyze network packages via debug containers",
	Long: `kubenet is a kubectl plugin that attaches debug containers to target pods
for network traffic analysis and monitoring in Kubernetes clusters.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kubenet - Network analysis kubectl plugin")
		fmt.Println("Use 'kubectl kubenet attach --help' for usage")
	},
}

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach debug container to target pod for network analysis",
	Long: `Attach a debug container to the specified pod to capture and analyze network traffic.
The debug container will have network monitoring tools installed.`,
	Run: func(cmd *cobra.Command, args []string) {
		if podName == "" {
			fmt.Fprintf(os.Stderr, "Error: pod name is required\n")
			os.Exit(1)
		}

		fmt.Printf("Attaching debug container to pod: %s\n", podName)
		if namespace != "" {
			fmt.Printf("Namespace: %s\n", namespace)
		}
		if tcpFilter != "" {
			fmt.Printf("TCP Filter: %s\n", tcpFilter)
		}

		// TODO: Implement debug container attachment
		attachDebugContainer()
	},
}

func init() {
	rootCmd.AddCommand(attachCmd)

	attachCmd.Flags().StringVarP(&podName, "pod", "p", "", "Target pod name (required)")
	attachCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Target namespace")
	attachCmd.Flags().StringVarP(&tcpFilter, "tcp-filter", "f", "", "TCP filter expression (e.g., 'port 80')")

	attachCmd.MarkFlagRequired("pod")
}

func attachDebugContainer() {
	fmt.Println("Creating debug container with network monitoring tools...")
	fmt.Printf("Target: %s/%s\n", namespace, podName)
	if tcpFilter != "" {
		fmt.Printf("Applying TCP filter: %s\n", tcpFilter)
	}

	manager, err := NewDebugContainerManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create debug container manager: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := manager.AttachDebugContainer(ctx, namespace, podName, tcpFilter); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach debug container: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}