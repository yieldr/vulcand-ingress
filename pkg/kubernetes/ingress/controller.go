package ingress

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/yieldr/vulcand-ingress/pkg/vulcan"
)

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	vulcan   *vulcan.Client
}

func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, vulcan *vulcan.Client) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		vulcan:   vulcan,
	}
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue.
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks
	// the key for other workers. This allows safe parallel processing because
	// two pods with the same key are never processed in parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic.
	err := c.syncToVulcan(key.(string))

	// Handle the error if something went wrong during the execution of the
	// business logic.
	c.handleErr(err, key)
	return true
}

// syncToVulcan is the business logic of the controller. In this controller it
// simply prints information about the pod to stdout. In case an error happened,
// it has to simply return the error. The retry logic should not be part of the
// business logic.
func (c *Controller) syncToVulcan(key string) error {
	item, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		return err
	}

	logger := log.WithField("ingress", key)

	if !exists {
		// Clean up all the entries in vulcan that relate to this ingress
		// resource.
		logger.Print("Ingress does not exist anymore")

		split := strings.Split(key, "/")
		if len(split) < 2 {
			logger.WithError(err).Printf("Key split failed for key %s", key)
			return err
		}

		ns := split[0]
		name := split[1]

		log.Debug("Deleting vulcan backend")
		if err := c.vulcan.DeleteBackend(ns, name); err != nil {
			log.WithError(err).Error("Failed deleting vulcan backend")
		}

		log.Debug("Deleting vulcan frontend")
		if err := c.vulcan.DeleteFrontend(ns, name); err != nil {
			log.WithError(err).Error("Failed deleting vulcan frontend")
		}

		log.Debug("Deleting vulcan server")
		if err := c.vulcan.DeleteServer(ns, name); err != nil {
			log.WithError(err).Error("Failed deleting vulcan server")
		}

	} else {
		// Note that you also have to check the uid if you have a local
		// controlled resource, which is dependent on the actual instance, to
		// detect that a Ingress was recreated with the same name.
		ingress := item.(*v1beta1.Ingress)

		// First we sync the ingresses default backend. This is a fallback
		// backend which will receive traffic if nothing else matches.
		backend := ingress.Spec.Backend

		logger.WithFields(log.Fields{
			"service": backend.ServiceName,
			"port":    backend.ServicePort.String(),
		}).Print("Syncing default ingress backend")

		c.vulcan.SyncBackend(ingress, backend)
		c.vulcan.SyncFrontend(ingress, backend, "", "")
		c.vulcan.SyncServer(ingress, backend)

		for _, rule := range ingress.Spec.Rules {

			for _, path := range rule.HTTP.Paths {

				logger = logger.WithFields(log.Fields{
					"host":    rule.Host,
					"path":    path.Path,
					"service": path.Backend.ServiceName,
					"port":    path.Backend.ServicePort.String(),
				})

				logger.Debug("Creating vulcan backend")
				if err := c.vulcan.SyncBackend(ingress, &path.Backend); err != nil {
					logger.WithError(err).Error("Failed creating vulcan backend")
				}

				logger.Debug("Creating vulcan frontend")
				if err := c.vulcan.SyncFrontend(ingress, &path.Backend, rule.Host, path.Path); err != nil {
					logger.WithError(err).Error("Failed creating vulcan frontend")
				}

				logger.Debug("Creating vulcan server")
				if err := c.vulcan.SyncServer(ingress, &path.Backend); err != nil {
					logger.WithError(err).Error("Failed creating vulcan server")
				}

				logger.Debug("Creating vulcan middleware")
				if err := c.vulcan.SyncMiddleware(ingress, backend); err != nil {
					logger.WithError(err).Error("Failed creating vulcan middleware")
				}
			}
		}
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every
		// successful synchronization. This ensures that future processing of
		// updates for this key is not delayed because of an outdated error
		// history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it
	// stops trying.
	if c.queue.NumRequeues(key) < 5 {
		log.Infof("Error syncing ingress %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later
		// again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could
	// not successfully process this key.
	runtime.HandleError(err)
	log.Infof("Dropping ingress %q out of the queue: %v", key, err)
}

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	log.Info("Starting ingress controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from
	// the queue is started.
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Info("Stopping ingress controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}
