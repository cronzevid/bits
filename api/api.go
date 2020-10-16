package api

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"
	"bytes"
	//"strconv"
	//"math"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"strings"
)

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("\n+++===%v took %v===+++\n\n", name, elapsed)
}

type Config struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
	Id     string `json:"id"`
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func sortedKeys(mapToSort map[string]string) []string {
	//keys := make([]string, 0, len(mapToSort))
	var keys []string

	for k, _ := range mapToSort {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	//res := make(map[string]string)
	//for _, k := range keys {
	//	res[k] = mapToSort[k]
	//}
        //log.Println(res)

	return keys
}

func PrepReq(config_data Config, api_url string, data map[string]string, req_type string) map[string]string {
	//defer TimeTrack(time.Now(), "prepare headers")

	xAuth := "BITSTAMP" + " " + config_data.Key

	xAuthNonce := fmt.Sprint(uuid.New())

	tmpstmp := makeTimestamp()
	xAuthTmpstmp := fmt.Sprintf("%d", tmpstmp)

	xAuthVersion := "v2"

	ContentType := ""

	host := strings.Split(api_url, "/")[0]
	path := "/" + strings.Join(strings.Split(api_url, "/")[1:], "/")

	var strData string
	dataSorted := sortedKeys(data)
        //log.Println(dataSorted)
	for _, key := range dataSorted {
            //log.Println(param)
	    strData = strData + key + "=" + string(data[key]) + "&"
	    ContentType = "application/x-www-form-urlencoded"
	}
	lastIndex := len(strData)
        // log.Println(strData)
	if lastIndex > 0 {
	    strData = strData[0 : lastIndex-1]
	}
        //log.Println(strData)

	var query string
	if strings.Contains(api_url, "?") {
            query = strings.Split(api_url, "?")[1]
	} else {
            query = ""
	}

	//log.Println("BITSTAMP" + " " + config_data.Key + req_type + host + path + query + ContentType + xAuthNonce + xAuthTmpstmp + xAuthVersion + strData)

	string_to_sign := "BITSTAMP " + config_data.Key + req_type + host + path + query + ContentType + xAuthNonce + xAuthTmpstmp + xAuthVersion + strData
	xAuthSig := GetHash(string_to_sign, config_data.Secret)
	//log.Println(xAuthSig)

	req := map[string]string{"X-Auth": xAuth, "X-Auth-Signature": xAuthSig, "X-Auth-Nonce": xAuthNonce, "X-Auth-Timestamp": xAuthTmpstmp, "X-Auth-Version": xAuthVersion, "Content-Type": ContentType}
	return req
}

func MakeReq(headers map[string]string, api_url string, method string, data map[string]string) string {
	//defer TimeTrack(time.Now(), "send request")

	postData := url.Values{}
	dataSorted := sortedKeys(data)
	for _, key := range dataSorted {
	    postData.Set(key, data[key])
	}

	client := &http.Client{}
	req, _ := http.NewRequest(method, "https://"+api_url, bytes.NewBufferString(postData.Encode()))
	for headerName, headerValue := range headers {
	    req.Header.Add(headerName, headerValue)
	}

	resp, _ := client.Do(req)

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	return string(bodyBytes)
}

func ApiMapAns(answer string) (map[string]string, error) {
	//json_ans := map[string]interface{}{}
	var json_ans map[string]string

	err := json.Unmarshal([]byte(answer), &json_ans)

	return json_ans, err

}

type M map[string]string
func ApiSliceAns(answer string) ([]M, error) {
	var json_ans []M

	err := json.Unmarshal([]byte(answer), &json_ans)

	return json_ans, err
}

func PostApiWrapper(config Config, api_url string, data map[string]string) (map[string]string, []M) {
	headers := PrepReq(config, api_url, data, "POST")
	ans := MakeReq(headers, api_url, "POST", data)

	var json_slice []M

	json_map, err := ApiMapAns(ans)
	if err != nil {
		json_slice, _ = ApiSliceAns(ans)
	}

	return json_map, json_slice
}

func GetApiWrapper(config Config, api_url string) (map[string]string, []M) {
	data := map[string]string{"": ""}
	headers := PrepReq(config, api_url, data, "GET")
	ans := MakeReq(headers, api_url, "GET", data)

	var json_slice []M

	json_map, err := ApiMapAns(ans)
	if err != nil {
		json_slice, _ = ApiSliceAns(ans)
	}

	return json_map, json_slice
}


func ReadJson(config string) Config {
	jsonFile, err := os.Open(config)
	defer jsonFile.Close()

	if err != nil {
		log.Println(err)
	}

	var res Config
	bytes, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(bytes, &res)

	return res
}

func GetHash(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return fmt.Sprintf("%X", h.Sum(nil)) //hex sig
}

//func main() {
//	config_data := ReadJson("btstmp.json")
//
//	postData := map[string]string{"": ""}
//	balance, _ := PostApiWrapper(config_data, "www.bitstamp.net/api/v2/balance/btcusd/", postData)
//	log.Println("-balance-")
//	log.Printf("BTC: %v, USD: %v\n", balance["btc_available"], balance["usd_available"])
//
//	postData = map[string]string{"offset": "0", "limit": "1", "sort": "desc"}
//	err, last_trade := PostApiWrapper(config_data, "www.bitstamp.net/api/v2/user_transactions/btcusd/", postData) // type 2 == market trade; -usd = buy, -btc = sell
//        if err != nil {
//	    log.Println(err)
//        }
//
//
//	log.Println("-last balance move-")
//        var action_type string
//	if last_trade[0]["type"] == "2" {
//            s, _ := strconv.ParseFloat(last_trade[0]["btc"], 32)
//            if math.Signbit(s) {
//                action_type = "Sold"
//	        log.Printf("%v %vUSD for %vBTC\n", action_type, last_trade[0]["usd"], last_trade[0]["btc"][1:len(last_trade[0]["usd"])])
//            } else {
//                action_type = "Bought"
//	        log.Printf("%v %vBTC for %vUSD\n", action_type, last_trade[0]["btc"], last_trade[0]["usd"][1:len(last_trade[0]["usd"])])
//            }
//        } else {
//	    log.Println(last_trade[0])
//        }
//
//	var config_data_new Config
//	log.Println("-last hour ticker-")
//	last_hour, _ := GetApiWrapper(config_data_new, "www.bitstamp.net/api/v2/ticker_hour/btcusd/")
//	open, _ := strconv.ParseFloat(last_hour["open"], 32)
//	close_, _ := strconv.ParseFloat(last_hour["last"], 32)
//        if math.Signbit(open - close_) {
//	    log.Println("Prise rose by", (1 - open/close_)*100)
//        } else {
//	    log.Println("Prise fell by", (1 - close_/open)*100)
//        }
//
//}
