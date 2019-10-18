package httpd

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"gopkg.in/mgo.v2"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
)

type K8sConfig struct {
	Clustername string `json:"clustername"`
	Configfile  string `json:"configfile"`
}

type K8sConfigs []K8sConfig

type HttpPod struct {
	ClusterCount int       `json:"clustercount"`
	PodNumber    int       `json:"pod_number"`
	Data         []PodStat `json:"data"`
}

type PodStat struct {
	Name       string `json:"name"`
	Cluster    string `json:"cluster"`
	Namespace  string `json:"namespace"`
	Status     string `json:"status"`
	Hostip     string `json:"hostip"`
	Msg        string `json:"msg"`
	Deployment string `json:"deployment"`
	//	Ready     bool   `json:"ready"`
}

func Podcheck(ClusterCfgs K8sConfigs) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		httpStat := HttpPod{}

		for _, ClusterCfg := range ClusterCfgs {
			httpStat.ClusterCount = len(ClusterCfgs)

			ch := make(chan []PodStat)

			go func(ClusterCfg K8sConfig) {
				postat := []PodStat{}
				config, err := clientcmd.RESTConfigFromKubeConfig([]byte(ClusterCfg.Configfile))
				if err != nil {
					Loges.Error("REST Config From KubeConfig is err:", zap.Error(err))
				}

				clientset, err := kubernetes.NewForConfig(config)
				if err != nil {
					Loges.Error("new KubeConfig is err:", zap.Error(err))
				}

				pods, err := clientset.CoreV1().Pods("kube-system").List(metav1.ListOptions{})
				if err != nil {
					Loges.Error("get pod info  is err:", zap.Error(err))
				}

				for _, po := range pods.Items { //for pod-status
					chpostat := make(chan PodStat)
					go podstat(chpostat, po, ClusterCfg.Clustername)
					v := <-chpostat
					close(chpostat)
					postat = append(postat, v)
				}
				ch <- postat
			}(ClusterCfg)

			vv := <-ch
			close(ch)
			httpStat.Data = append(httpStat.Data, vv...)

			//config, err := clientcmd.RESTConfigFromKubeConfig([]byte(ClusterCfg.Configfile))
			//if err != nil {
			//	fmt.Println("6666666666666666", err)
			//}
			//
			//clientset, err := kubernetes.NewForConfig(config)
			//if err != nil {
			//	log.Fatalln(err)
			//}
			//
			//pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
			//if err != nil {
			//	log.Println(err.Error())
			//}
			//
			//httpStat.ClusterCount = len(ClusterCfgs)
			//
			//for _, po := range pods.Items { //for pod-status
			//	chpostat := make(chan PodStat)
			//	go podstat(chpostat, po, ClusterCfg.Clustername)
			//	v := <-chpostat
			//	close(chpostat)
			//	httpStat.Data = append(httpStat.Data, v)
			//}

		}

		//do alter
		alter(httpStat.Data)

		httpStat.PodNumber = len(httpStat.Data)

		w2, _ := json.Marshal(httpStat)
		w.Write(w2)

	}
}

func podstat(chpostat chan PodStat, po v1.Pod, Clustername string) {
	s1 := PodStat{}
	for _, pds := range po.Status.ContainerStatuses {
		if pds.Ready != true {
			//if pds.Name == "trace-test" /*|| pds.Name == "imcp-web" */ {
			s1.Msg = "Yellow"
		}

	}
	s1.Name = po.Name
	s1.Namespace = po.Namespace
	s1.Status = string(po.Status.Phase)
	s1.Cluster = Clustername
	s1.Hostip = po.Status.HostIP
	s1.Deployment = strings.Split(po.Name, "-dpt-")[0]

	chpostat <- s1

}

func alter(pos []PodStat) {
	ss := "podName : "
	alterList := make(map[string][]PodStat)
	for _, j := range pos {

		if j.Msg == "Yellow" {
			alterList[j.Deployment] = append(alterList[j.Deployment], j)
		}

	}

	for k, v := range alterList {
		for _, j := range v {
			ss = ss + j.Name + ","
		}
		//fmt.Println("PodCheck", "project: "+k, "cluster: "+v[0].Cluster, "ns: "+v[0].Namespace, ss)
		go SendWechat("PodCheck", "project: "+k, "cluster: "+v[0].Cluster, "ns: "+v[0].Namespace, ss)
	}

}

func GetConfig() K8sConfigs {
	session, err := mgo.Dial(viper.GetString("db.addr"))
	if err != nil {
		Loges.Error("conn mgo is err:", zap.Error(err))
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	err = session.DB("admin").Login(viper.GetString("db.dbuser"), viper.GetString("db.dbpass"))
	if err != nil {
		Loges.Error("auth mgo is err:", zap.Error(err))
	}
	aa := K8sConfigs{}
	c := session.DB("check").C("k8sconfig")
	err = c.Find(nil).All(&aa)
	if err != nil {
		Loges.Error("select db is err:", zap.Error(err))
	}

	return aa
}

//func AddConfig(){
//	bb,err := ioutil.ReadFile("/home/han/config")
//	if err !=nil {
//		fmt.Println("111111111",err)
//	}
//	k8sc := K8sConfig{}
//	k8sc.Clustername = ""
//	k8sc.Configfile = string(bb)
//
//	//fmt.Println(k8sc)
//
//
//	session,err := mgo.Dial(viper.GetString("db.addr"))
//	if err !=nil {
//		fmt.Println("333333333333333333333",err)
//	}
//	defer session.Close()
//	session.SetMode(mgo.Monotonic, true)
//	err = session.DB("admin").Login(viper.GetString("db.dbuser"),viper.GetString("db.dbpass"))
//	if err !=nil{
//		fmt.Println("2222222222222222222",err)
//	}
//	c := session.DB("check").C("k8sconfig")
//	err = c.Insert(&k8sc)
//	if err !=nil{
//		fmt.Println("44444444444444444444",err)
//	}
//}
