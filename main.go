package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    
    "k8s.io/apimachinery/pkg/fields"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/cache"
    "k8s.io/client-go/util/homedir"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "path/filepath"
    corev1 "k8s.io/api/core/v1"
)

func main() {
    // Define a flag for the kubeconfig path
    var kubeconfig string
    flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "Path to a kubeconfig file")
    flag.Parse() // Parse all flags

    var config *rest.Config
    var err error

    // Decide whether to use in-cluster config or kubeconfig based on the existence of the file at kubeconfig path
    if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
        fmt.Println("Using in-cluster config")
        config, err = rest.InClusterConfig()
        if err != nil {
            panic(err.Error())
        }
    } else {
        fmt.Println("Using kubeconfig file:", kubeconfig)
        config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
            panic(err.Error())
        }
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }

    // Setting up signal handling
    stopCh := make(chan struct{})
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", corev1.NamespaceAll, fields.Everything())
    _, controller := cache.NewInformer(watchlist, &corev1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
        AddFunc: func(obj interface{}) {
            pod := obj.(*corev1.Pod)
            fmt.Printf("Pod added: %s \n", pod.Name)
        },
    })

    go func() {
        <-sigs
        fmt.Println("Received termination signal, shutting down...")
        close(stopCh)
    }()

    controller.Run(stopCh)
}

