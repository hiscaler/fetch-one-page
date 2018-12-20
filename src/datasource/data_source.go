package datasource

import (
	"io/ioutil"
	"log"
	"net/http"
	"config"
	"fmt"
)

type IDataSource interface {
	Read() ([]byte, bool)
}

// Local Data Source
type LocalDataSource struct {
}

func (ds *LocalDataSource) Read() ([]byte, bool) {
	content, err := ioutil.ReadFile("data/feed.json")
	if err != nil {
		log.Println(err)
		return nil, false
	}

	return content, true
}

// API Data Source
type ApiDataSource struct {
}

func (ds *ApiDataSource) Read() ([]byte, bool) {
	cfg := *config.NewConfig()
	resp, err := http.Get(cfg.ApiEndpoint + "/url")
	if err != nil {
		fmt.Println(err)
		return nil, false
	}
	defer resp.Body.Close()
	if respBody, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, false
	} else {
		return respBody, true
	}
}
