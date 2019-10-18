package httpd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"

	"github.com/julienschmidt/httprouter"

	"github.com/spf13/viper"
)

type EsClusterStat struct {
	ClusterName                 string  `json:"cluster_name"`
	Status                      string  `json:"status"`
	TimeOut                     bool    `json:"time_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float32 `json:"active_shards_percent_as_number"`
}

type EsStatus struct {
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

type HttpEs struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data []EsStatus `json:"data"`
}

func Escheck(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	httpStat := HttpEs{}
	esList := viper.GetStringSlice("es.eslist")
	for _, ip := range esList {
		ch := make(chan EsClusterStat)
		go func(ip string) {
			url := "http://" + ip + "/_cluster/health?pretty"

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				Loges.Error("new request is err:", zap.Error(err))
			}

			req.SetBasicAuth(viper.GetString("es.es-user"), viper.GetString("es.es-pwd"))
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				Loges.Error("do request is err:", zap.Error(err))

				ch <- EsClusterStat{Status: "red"}
				return
			}
			defer res.Body.Close()

			resData := EsClusterStat{}
			body, _ := ioutil.ReadAll(res.Body)
			err = json.Unmarshal(body, &resData)
			if err != nil {
				Loges.Error("json exchange  is err:", zap.Error(err))
			}

			ch <- resData
		}(ip)

		v := <-ch

		if v.Status == "red" {
			httpStat.Msg = "yellow"
			// SendWechat(ip)
			httpStat.Data = append(httpStat.Data, EsStatus{Msg: "red", Data: ip})
		} else if v.Status == "yellow" {
			httpStat.Msg = "yellow"
			httpStat.Data = append(httpStat.Data, EsStatus{Msg: "yellow", Data: ip})
		} else if v.NumberOfNodes != viper.GetInt("es.nodes") {
			httpStat.Msg = "yellow"
			//	SendWechat(ip)
			httpStat.Data = append(httpStat.Data, EsStatus{Msg: "yellow", Data: ip})
		} else {
			httpStat.Code = 200
			httpStat.Msg = "green"
			httpStat.Data = append(httpStat.Data, EsStatus{Msg: "green", Data: ip})
		}
	}
	w2, _ := json.Marshal(httpStat)
	w.Write(w2)
}
