package api

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"
	"bytes"
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
	var keys []string

	for k, _ := range mapToSort {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func PrepReq(config_data Config, api_url string, data map[string]string, req_type string) map[string]string {

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
	for _, key := range dataSorted {
	    strData = strData + key + "=" + string(data[key]) + "&"
	    ContentType = "application/x-www-form-urlencoded"
	}
	lastIndex := len(strData)
	if lastIndex > 0 {
	    strData = strData[0 : lastIndex-1]
	}

	var query string
	if strings.Contains(api_url, "?") {
            query = strings.Split(api_url, "?")[1]
	} else {
            query = ""
	}

	string_to_sign := "BITSTAMP " + config_data.Key + req_type + host + path + query + ContentType + xAuthNonce + xAuthTmpstmp + xAuthVersion + strData
	xAuthSig := GetHash(string_to_sign, config_data.Secret)

	req := map[string]string{"X-Auth": xAuth, "X-Auth-Signature": xAuthSig, "X-Auth-Nonce": xAuthNonce, "X-Auth-Timestamp": xAuthTmpstmp, "X-Auth-Version": xAuthVersion, "Content-Type": ContentType}
	return req
}

func MakeReq(headers map[string]string, api_url string, method string, data map[string]string) string {

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

