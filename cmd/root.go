package cmd

import (
	"fmt"
	"os"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/spf13/cobra"

	"github.com/yieldr/vulcand-ingress/pkg/kubernetes"
	"github.com/yieldr/vulcand-ingress/pkg/kubernetes/ingress"
	"github.com/yieldr/vulcand-ingress/pkg/vulcan"
)

var cmdRoot = &cobra.Command{
	Use:   "vulcand-ingress",
	Short: "vulcand ingress controller",
	Long: `Vulcand Ingress is a Kubernetes Ingress Controller for the Vulcand
reverse proxy.`,
	Run: runRoot,
}

func runRoot(cmd *cobra.Command, args []string) {

	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	clientset, err := kubernetes.New(kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed creating kubernetes client. %s", err)
		os.Exit(1)
	}

	vulcanAddr, _ := cmd.Flags().GetString("vulcand-addr")
	vulcan := vulcan.New(vulcanAddr)

	namespace, _ := cmd.Flags().GetString("namespace")

	selector, _ := cmd.Flags().GetString("selector")
	fields, err := fields.ParseSelector(selector)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid selector. %s", err)
		os.Exit(1)
	}

	ingressWatcher := cache.NewListWatchFromClient(clientset.ExtensionsV1beta1().RESTClient(), "ingresses", namespace, fields)

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

	controller := ingress.NewController(queue, indexer, informer, vulcan)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	select {}
}

func init() {
	cmdRoot.Flags().String("kubeconfig", "", "Absolute path to the kubeconfig file. If empty an in-cluster configuration is assumed.")
	cmdRoot.Flags().String("namespace", "", "Namespace in which to watch for resources.")
	cmdRoot.Flags().String("selector", "", "Selector with which to match resources.")
	cmdRoot.Flags().String("vulcand-addr", "http://localhost:8182", "Vulcand API address.")
}

func Execute() {
	if err := cmdRoot.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
