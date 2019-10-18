package httpd

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
)

type HttpNode struct {
	ClusterCount int        `json:"clustercount"`
	NodeNumber   int        `json:"node_number"`
	Data         []NodeStat `json:"data"`
}

type NodeStat struct {
	Name    string `json:"name"`
	Cluster string `json:"cluster"`
	Status  string `json:"status"`
}

func NodeCheck(ClusterCfgs K8sConfigs) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		httpStat := HttpNode{}

		nodestat := NodeStat{}

		nodeList := v1.NodeList{}.Items

		for _, ClusterCfg := range ClusterCfgs {
			httpStat.ClusterCount = len(ClusterCfgs)

			config, err := clientcmd.RESTConfigFromKubeConfig([]byte(ClusterCfg.Configfile))
			if err != nil {
				Loges.Error("REST Config From KubeConfig is err:", zap.Error(err))
			}

			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				Loges.Error("new KubeConfig is err:", zap.Error(err))
			}
			nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
			if err != nil {
				Loges.Error("get nodes info  is err:", zap.Error(err))
			}

			nodeList = append(nodeList, nodes.Items...)
		}
		for _, i := range nodeList {
			ch := make(chan NodeStat)
			go func(i v1.Node) {
				if i.Status.Conditions[len(i.Status.Conditions)-1].Type != "Ready" {
					nodestat.Status = "NotReady"
					//	i.ClusterName = ClusterCfgs[1].Clustername
					nodestat.Name = i.Name
				}

				nodestat.Status = "Ready"
				//nodestat.Cluster = i.ClusterName
				nodestat.Name = i.Name

				ch <- nodestat

			}(i)
			nodest := <-ch
			close(ch)

			if nodest.Status != "Ready" {
				//if nodest.Name == "10.180.8.8" {
				err := SendWechat("NodeCheck", nodest.Name, nodest.Status)
				if err != nil {
					Loges.Error("send msg is err", zap.Error(err))
				}
			}

			httpStat.Data = append(httpStat.Data, nodest)
			httpStat.NodeNumber = len(httpStat.Data)
		}
		w2, _ := json.Marshal(httpStat)
		w.Write(w2)

	}
}
