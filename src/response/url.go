package response

import (
	"log"
	"strings"
	"net/http"
	"time"
	"io/ioutil"
	"net/url"
	"errors"
)

// 单条爬取地址
type Url struct {
	Id              string
	Url             string
	Status          int
	Project         Project
	CallbackMessage string
}

// 回调
func (this *Url) Callback(endpoint string) error {
	params := url.Values{}
	params.Add("status", string(this.Status))
	params.Add("message", string(this.CallbackMessage))
	payload := strings.NewReader(params.Encode())
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	req, err := http.NewRequest("PATCH", endpoint+"/callback/"+this.Id, payload)
	if err != nil {
		log.Fatalln("Request error: " + err.Error())
		return err
	}
	req.Body.Close()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalln("Response error: ", err)
		return err
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		log.Println("Callback is successful." + string(respBody))
		return nil
	} else {
		return errors.New("Callback is failed, http code is " + resp.Status)
	}
}
