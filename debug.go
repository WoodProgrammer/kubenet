package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type DebugContainerManager struct {
	clientset *kubernetes.Clientset
}

func NewDebugContainerManager() (*DebugContainerManager, error) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &DebugContainerManager{clientset: clientset}, nil
}

func (d *DebugContainerManager) AttachDebugContainer(ctx context.Context, namespace, podName, tcpFilter string) error {
	fmt.Printf("Fetching pod %s in namespace %s...\n", podName, namespace)

	pod, err := d.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	debugContainerName := "kubenet-debug"

	debugContainer := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:  debugContainerName,
			Image: "nicolaka/netshoot:latest",
			Command: []string{
				"/bin/bash",
				"-c",
				"echo 'Debug container attached. Use tcpdump, ss, netstat for network analysis.' && sleep 3600",
			},
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{
						"NET_ADMIN",
						"NET_RAW",
					},
				},
			},
			Stdin: true,
			TTY:   true,
		},
	}

	if tcpFilter != "" {
		debugContainer.Command = []string{
			"/bin/bash",
			"-c",
			fmt.Sprintf("echo 'Debug container with TCP filter: %s' && tcpdump -i any %s & sleep 3600", tcpFilter, tcpFilter),
		}
	}

	if pod.Spec.EphemeralContainers == nil {
		pod.Spec.EphemeralContainers = []corev1.EphemeralContainer{}
	}

	for _, ec := range pod.Spec.EphemeralContainers {
		if ec.Name == debugContainerName {
			return fmt.Errorf("debug container %s already exists", debugContainerName)
		}
	}

	pod.Spec.EphemeralContainers = append(pod.Spec.EphemeralContainers, debugContainer)

	fmt.Printf("Creating ephemeral debug container '%s'...\n", debugContainerName)

	_, err = d.clientset.CoreV1().Pods(namespace).UpdateEphemeralContainers(
		ctx, podName, pod, metav1.UpdateOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to create debug container: %w", err)
	}

	fmt.Printf("✓ Debug container '%s' attached successfully\n", debugContainerName)
	fmt.Printf("To exec into the debug container, run:\n")
	fmt.Printf("kubectl exec -it %s -c %s -n %s -- /bin/bash\n", podName, debugContainerName, namespace)

	return nil
}