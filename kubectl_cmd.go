package etcdio

import (
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/runtime"
	//labelsutil "k8s.io/kubernetes/pkg/util/labels"
	"k8s.io/kubernetes/pkg/apis/extensions"
	utilruntime "k8s.io/kubernetes/pkg/util/runtime"
	"k8s.io/kubernetes/pkg/watch"
)

type EtcdIo struct {
	client     clientset.Interface
	podControl controller.PodControlInterface

	APIServerList []string

	Hostname string

	WaitTime int

	// A store of pods, populated by the podController
	podStore cache.StoreToPodLister
	// Watches changes to all pods
	podController *framework.Controller
	// podStoreSynced returns true if the pod store has been synced at least once.
	// Added as a member to the struct to allow injection for testing.
	podStoreSynced func() bool

	podStatusLock sync.Mutex

	podStatusMap map[string]*api.Pod

	//Micro-BenchMark sequence
	mbmSeq int

	WriteLogFlag bool
	WriteLogPath string
}

func (ei *EtcdIo) Initial() error {
	glog.Info("testclient startted")

	kubeconfig, err := clientcmd.BuildConfigFromFlags(ei.APIServerList[0], "")
	if err != nil {
		return err
	}
	kubeconfig.QPS = 20.0
	kubeconfig.Burst = 30

	ei.podStatusMap = make(map[string]*api.Pod)

	ei.client = clientset.NewForConfigOrDie(restclient.AddUserAgent(kubeconfig, "testclient-etcdio"))

	resyncPeriod := func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(float64(12*time.Hour.Nanoseconds()) * factor)
	}

	ei.podStore.Indexer, ei.podController = framework.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return ei.client.Core().Pods(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return ei.client.Core().Pods(api.NamespaceAll).Watch(options)
			},
		},
		&api.Pod{},
		resyncPeriod(),
		framework.ResourceEventHandlerFuncs{
			AddFunc:    ei.addPod,
			UpdateFunc: ei.updatePod,
			DeleteFunc: ei.deletePod,
		},
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	ei.podStoreSynced = ei.podController.HasSynced
	pathLen := len(ei.WriteLogPath)
	if ei.WriteLogFlag && pathLen > 0 {
		if _, err := os.Stat(ei.WriteLogPath); err != nil {
			err = os.MkdirAll(ei.WriteLogPath, 0777)
			if err != nil {
				ei.WriteLogPath = ""
			}
		}
		if ei.WriteLogPath[pathLen-1] != '/' {
			ei.WriteLogPath = ei.WriteLogPath + "/"
		}
	}
	return nil
}

func (ei *EtcdIo) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&ei.APIServerList, "api-servers", []string{}, "List of Kubernetes API servers for publishing events, and reading pods and services. (ip:port), comma separated.")
	fs.StringVar(&ei.Hostname, "hostname", ei.Hostname, "If non-empty, will use this string as identification instead of the actual hostname.")
	fs.IntVar(&ei.WaitTime, "wait-time", 8, "wait duration before test begin")
	fs.BoolVar(&ei.WriteLogFlag, "write-log", false, "Enable Write QPS Log")
	fs.StringVar(&ei.WriteLogPath, "log-path", ei.WriteLogPath, "QPS log writed in this Path")
}

func (ei *EtcdIo) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	go ei.podController.Run(stopCh)

	go ei.TestClientWork()

	<-stopCh
	glog.Infof("Shutting down etcdio\n")
}

// When a pod is created, ensure its controller syncs
func (ei *EtcdIo) addPod(obj interface{}) {
	pod, ok := obj.(*api.Pod)
	if !ok || !ei.isTestPod(pod) {
		return
	}
	if ei.isRunning(pod) && ei.isTestPodForCreate(pod) {
		ei.podStatusLock.Lock()
		ei.podStatusMap[string(pod.ObjectMeta.UID)] = pod
		ei.podStatusLock.Unlock()
		glog.Infof("Pod %s created\n", pod.Name)
	}
}

// updatePod figures out what deployment(s) manage the ReplicaSet that manages the Pod when the Pod
// is updated and wake them up. If anything of the Pods have changed, we need to awaken both
// the old and new deployments. old and cur must be *api.Pod types.
func (ei *EtcdIo) updatePod(old, cur interface{}) {
	if api.Semantic.DeepEqual(old, cur) {
		return
	}
	curPod := cur.(*api.Pod)
	oldPod := old.(*api.Pod)
	if !ei.isTestPod(curPod) {
		return
	}
	if ei.isPending(oldPod) && ei.isRunning(curPod) {
		// wait before test start
		time.Sleep(time.Duration(ei.WaitTime) * time.Second)
		ei.podStatusLock.Lock()
		ei.podStatusMap[string(curPod.ObjectMeta.UID)] = curPod
		ei.podStatusLock.Unlock()
		glog.Infof("Pod %s Running \n", curPod.Name)
	}

}

// When a pod is deleted, ensure its controller syncs.
// obj could be an *api.Pod, or a DeletionFinalStateUnknown marker item.
func (ei *EtcdIo) deletePod(obj interface{}) {
	pod, ok := obj.(*api.Pod)

	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			glog.Errorf("Couldn't get object from tombstone %+v", obj)
			return
		}
		pod, ok = tombstone.Obj.(*api.Pod)
		if !ok {
			glog.Errorf("Tombstone contained object that is not a pod %+v", obj)
			return
		}
	}
	ei.podStatusLock.Lock()
	delete(ei.podStatusMap, string(pod.ObjectMeta.UID))
	ei.podStatusLock.Unlock()
	glog.V(4).Infof("Pod %s deleted\n", pod.Name)

}

func (ei *EtcdIo) UpdateScaleup(su *extensions.ScaleUpMaxQps) (*extensions.ScaleUpMaxQps, error) {
	return ei.client.Extensions().ScaleUpMaxQpses(su.ObjectMeta.Namespace).Update(su)
}
func (ei *EtcdIo) UpdateScaleUpDelay(su *extensions.ScaleUpDelay) (*extensions.ScaleUpDelay, error) {
	return ei.client.Extensions().ScaleUpDelaies(su.ObjectMeta.Namespace).Update(su)
}

func (ei *EtcdIo) UpdateSvcInterference(su *extensions.SvcInterference) (*extensions.SvcInterference, error) {
	return ei.client.Extensions().SvcInterferences(su.ObjectMeta.Namespace).Update(su)
}

func (ei *EtcdIo) UpdateScaleOutMaxQps(su *extensions.ScaleOutMaxQps) (*extensions.ScaleOutMaxQps, error) {
	return ei.client.Extensions().ScaleOutMaxQpses(su.ObjectMeta.Namespace).Update(su)
}
func (ei *EtcdIo) UpdateScaleOutDelay(su *extensions.ScaleOutDelay) (*extensions.ScaleOutDelay, error) {
	return ei.client.Extensions().ScaleOutDelaies(su.ObjectMeta.Namespace).Update(su)
}

func (ei *EtcdIo) UpdateTcStatus(su *extensions.TcStatus) (*extensions.TcStatus, error) {
	return ei.client.Extensions().TcStatuses(su.ObjectMeta.Namespace).Update(su)
}

func (ei *EtcdIo) GetScaleup(namespace, name string) (*extensions.ScaleUpMaxQps, error) {
	return ei.client.Extensions().ScaleUpMaxQpses(namespace).Get(name)
}

func (ei *EtcdIo) GetScaleUpDelay(namespace, name string) (*extensions.ScaleUpDelay, error) {
	return ei.client.Extensions().ScaleUpDelaies(namespace).Get(name)
}

func (ei *EtcdIo) GetSvcInterference(namespace, name string) (*extensions.SvcInterference, error) {
	return ei.client.Extensions().SvcInterferences(namespace).Get(name)
}

func (ei *EtcdIo) GetScaleOutMaxQps(namespace, name string) (*extensions.ScaleOutMaxQps, error) {
	return ei.client.Extensions().ScaleOutMaxQpses(namespace).Get(name)
}

func (ei *EtcdIo) GetScaleOutDelay(namespace, name string) (*extensions.ScaleOutDelay, error) {
	return ei.client.Extensions().ScaleOutDelaies(namespace).Get(name)
}

func (ei *EtcdIo) GetTcStatus(namespace, name string) (*extensions.TcStatus, error) {
	return ei.client.Extensions().TcStatuses(namespace).Get(name)
}

func (ei *EtcdIo) DeleteTcStatus(namespace, name string) error {
	return ei.client.Extensions().TcStatuses(namespace).Delete(name, nil)
}

func (ei *EtcdIo) UpdateProfileStatus(pf *api.Profile) (*api.Profile, error) {
	return ei.client.Core().Profiles(pf.ObjectMeta.Namespace).UpdateStatus(pf)
}

func (ei *EtcdIo) GetNode(name string) (*api.Node, error) {
	return ei.client.Core().Nodes().Get(name)
}

func (ei *EtcdIo) isRunning(pod *api.Pod) bool {
	return string(pod.Status.Phase) == "Running"
}
func (ei *EtcdIo) isPending(pod *api.Pod) bool {
	return string(pod.Status.Phase) == "Pending"
}

func (ei *EtcdIo) isTestPod(pod *api.Pod) bool {

	_, ok := pod.ObjectMeta.Labels["testDimensions"]
	pf, err := ei.getMatchProfile(pod)
	if err != nil || pf.Spec.Profile.AppClass != "service" {
		return false
	}

	if pod.Spec.NodeName == ei.Hostname && ok {
		return true
	}
	return false
}

func (ei *EtcdIo) isTestPodForCreate(pod *api.Pod) bool {
	_, ok := pod.ObjectMeta.Labels["testDimensions"]
	pf, err := ei.getMatchProfile(pod)
	if err != nil || pf.Spec.Profile.AppClass != "service" {
		return false
	}

	if pod.Spec.NodeName == ei.Hostname && ok {
		var tmpName string
		switch pod.ObjectMeta.Labels["testDimensions"] {
		case "scale-up":
			tmpName = pod.ObjectMeta.Labels["profile"] + "_" + pod.ObjectMeta.Labels["testDimensions"] + "_" + pod.ObjectMeta.Labels["podSize"]
		case "interference":
			tmpName = pod.ObjectMeta.Labels["profile"] + "_" + pod.ObjectMeta.Labels["testDimensions"]
		case "scale-out":
			tmpName = pod.ObjectMeta.Labels["profile"] + "_" + pod.ObjectMeta.Labels["testDimensions"] + "_" + pod.ObjectMeta.Labels["replicas"]
		default:
			return false
		}
		if status, ok := pf.Status.TestPodStatuses[tmpName]; ok {
			if status.State != api.TestFinished && status.State != api.TestFailed {
				return true
			}
		}
	}
	return false
}
