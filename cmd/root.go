package cmd

import (
	"fmt"
	"os"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/spf13/cobra"

	"github.com/yieldr/vulcand-ingress/pkg/ingress"
)

var cmdRoot = &cobra.Command{
	Use:   "vulcand-ingress",
	Short: "vulcand ingress controller",
	Long: `Vulcand Ingress is a Kubernetes Ingress Controller for the Vulcand
reverse proxy.`,
	Run: runRoot,
}

func runRoot(cmd *cobra.Command, args []string) {

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed retrieving in-cluster configuration. %s", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed creating kubernetes client. %s", err)
		os.Exit(1)
	}

	ingressWatcher := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"ingress",
		v1.NamespaceAll,
		fields.Everything())

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	resourceHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}

	indexer, informer := cache.NewIndexerInformer(
		ingressWatcher,
		&v1beta1.Ingress{},
		0,
		resourceHandler,
		cache.Indexers{})

	controller := ingress.NewController(queue, indexer, informer)

	indexer.Add(&v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myingress",
			Namespace: v1.NamespaceDefault,
		},
	})

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	select {}
}

func init() {
	cmdRoot.Flags().String("kubeconfig", "", "absolute path to the kubeconfig file")
}

func Execute() {
	if err := cmdRoot.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
