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
    var kubeconfig *string
    if home := homedir.HomeDir(); home != "" {
        kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
    } else {
        kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
    }
    flag.Parse()

    var config *rest.Config
    var err error
    if *kubeconfig == "" {
        config, err = rest.InClusterConfig()
    } else {
        config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
    }
    if err != nil {
        panic(err.Error())
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

