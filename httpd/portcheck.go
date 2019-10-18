package httpd

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/julienschmidt/httprouter"

	"github.com/spf13/viper"
)

type PortStatus struct {
	Msg   string `json:"msg"`
	Data  string `json:"data"`
	Topic string `json:"topic"`
}

type HttpStat struct {
	Number int          `json:"number"`
	Msg    string       `json:"msg"`
	Data   []PortStatus `json:"data"`
}

type SearchMachine struct {
	Httpstatus int `json:"httpstatus"`
	Data       map[string][]struct {
		Hostname string `json:"hostname"`
		Ip       string `json:"ip"`
		Sn       string `json:"sn"`
	}
}

func PostPort() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		t, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Loges.Error("error : ", zap.Error(err))
		}
		var port1 PortStatus
		err = json.Unmarshal(t, &port1)
		if err != nil {
			Loges.Error("format json error : ", zap.Error(err))
		}
		c := session.DB("check").C("checkport")
		err = c.Insert(&port1)
		if err != nil {
			Loges.Error("insert db error : ", zap.Error(err))
		}
	}

}

func GetPort() []PortStatus {
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

	var aa []PortStatus
	c := session.DB("check").C("checkport")
	err = c.Find(nil).All(&aa)
	if err != nil {
		Loges.Error("select db is err:", zap.Error(err))
	}
	//fmt.Println(aa)
	//var portList []string
	//for _, j := range aa {
	//	portList = append(portList, j.Data)
	//}

	return aa

}

func PortCheck(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	httpStat := HttpStat{}
	//portList := viper.GetStringSlice("check.ipPort")
	portList := GetPort()

	for _, node := range portList {
		ch := make(chan PortStatus)
		go func(ip string) {
			//s1 := PortStatus{}
			//s1.Data = ip
			conn, err := net.Dial("tcp", ip)
			if err != nil {
				Loges.Error("port is err", zap.Error(err))
				node.Msg = "red"
			} else {
				node.Msg = "green"
				conn.Close()
			}

			ch <- node

		}(node.Data)

		v := <-ch
		close(ch)
		if v.Msg == "red" {
			fmt.Println(v.Data)
			hostname := GetHostNname(v.Data)

			err := SendWechat("PortCheck", v.Topic, v.Data, hostname)
			if err != nil {
				Loges.Error("send msg is err", zap.Error(err))
			}
			httpStat.Msg = "yellow"
		}
		httpStat.Data = append(httpStat.Data, v)
	}

	if httpStat.Msg != "yellow" {
		httpStat.Msg = "green"
	}

	httpStat.Number = len(httpStat.Data)
	w2, _ := json.Marshal(httpStat)
	w.Write(w2)

}

func GetHostNname(iport string) string {
	ss := strings.Split(iport, ":")
	ip := ss[0]
	url := "https://registry.monitor.ifengidc.com/api/v1/resource/search?ns=loda&type=machine&v=" + ip
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Loges.Error("new request is err", zap.Error(err))
	}

	req.Header.Add("Authtoken", "hanlei3:NwMiBoSpDbEJYdl5mVLJ")
	req.Header.Add("NS", "loda")
	req.Header.Add("Resource", "machine")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		Loges.Error("do request is err", zap.Error(err))
	}
	defer res.Body.Close()
	s, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Loges.Error("read res.body  is err:", zap.Error(err))
	}

	kk := SearchMachine{}
	err = json.Unmarshal(s, &kk)
	if err != nil {
		Loges.Error("json exchange is err", zap.Error(err))
	}

	for k, _ := range kk.Data {
		return kk.Data[k][0].Hostname
	}

	return "no hostname"

}

//func ClientGet(url string, seconds time.Duration) httprouter.Handle {
//	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//		for {
//			req, err := http.NewRequest("GET", url, nil)
//			if err != nil {
//				fmt.Println("client get func is err : ", err)
//			}
//			client := http.Client{}
//			res, err := client.Do(req)
//			if err != nil {
//				fmt.Println("client get func is err2 ", err)
//			}
//
//			defer res.Body.Close()
//
//			sleepTime := time.Second * seconds
//
//			time.Sleep(sleepTime)
//
//		}
//		//w2, _ := json.Marshal(HttpBody{code: 200, Msg: "ok", Data: ""})
//		//w.Write(w2)
//
//	}
//}
