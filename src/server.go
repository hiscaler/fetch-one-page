package main

import (
	"io/ioutil"
	"log"
	"response"
	"encoding/json"
	"fmt"
	"strings"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"time"
	"net/url"
	"config"
	"sync"
	"datasource"
	"strconv"
)

var cfg config.Config
var wg sync.WaitGroup

func init() {
	cfg = *config.NewConfig()
	log.SetFlags(log.Ldate | log.Lshortfile)
}

func Read() ([]byte, bool) {
	content, err := ioutil.ReadFile("data/feed.json")
	if err != nil {
		log.Fatalln(err)
		return nil, false
	}

	return content, true
}


func FetchOne(rUrl response.Url, wg *sync.WaitGroup) {
	wg.Done()
	urlPath := rUrl.Url
	log.Println("> " + urlPath)
	project := rUrl.Project
	msg := "> Page Render method: " + project.PageRenderMethod + ", Use Agent: "
	if project.UseAgent {
		msg += "Yes"
	} else {
		msg += "No"
	}
	resp, err := http.Get(urlPath)
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode == 200 {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		} else {
			// Parse page source code
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
			if err != nil {
				log.Fatal(err)
			}
			cellsRawData := make(map[string][]string, 0)
			for _, prop := range project.Props {
				propValues := []string{}
				text := ""
				for _, rule := range prop.Rules {
					if len(text) > 0 {
						break
					}
					if rule.RuleType == "css" {
						// Find single item
						//switch rule.Parser {
						//case "text":
						//	text = doc.Find(rule.Path).Text()
						//case "raw":
						//	if html, err := doc.Find(rule.Path).Html(); err == nil {
						//		text = html
						//	} else {
						//		log.Println(err)
						//	}
						//case "attr":
						//	text, _ = doc.Find(rule.Path).Attr(rule.Attr)
						//}
						// Find multiple items
						doc.Find(rule.Path).Each(func(i int, s *goquery.Selection) {
							switch rule.Parser {
							case "text":
								text = s.Text()
							case "raw":
								if html, err := s.Html(); err == nil {
									text = html
								} else {
									log.Println(err)
								}
							case "attr":
								text, _ = s.Attr(rule.Attr)
							}

							propValues = append(propValues, text)
						})
					} else if rule.RuleType == "xpath" {
						// todo
						log.Println("Parse type is XPath, Don't implemented.")
					}
				}
				cellsRawData[prop.Name] = propValues
			}
			log.Println(fmt.Sprintf("%#v", cellsRawData))

			cells := make([]map[string]string, 0)
			if len(cellsRawData) > 0 {
				usedPropName := ""
				minSize := 1024
				for name, items := range cellsRawData {
					l := len(items)
					if l != 0 && l < minSize {
						minSize = l
						usedPropName = name
					} else if minSize == 1 {
						break
					}
				}

				for k, v := range cellsRawData[usedPropName] {
					cellValue := map[string]string{
						usedPropName: v,
					}
					for _, prop := range project.Props {
						if prop.Name != usedPropName {
							cellValue[prop.Name] = cellsRawData[prop.Name][k]
						}
					}
					log.Println(fmt.Sprintf("%#v", cellValue))
					cells = append(cells, cellValue)
				}
			}
			log.Println(fmt.Sprintf("%#v", cells))

			for i, cell := range cells {
				log.Println("Callback #" + strconv.Itoa(i) + " Data")
				params := url.Values{}
				params.Add("url_id", rUrl.Id)
				cellJson, _ := json.Marshal(cell)
				params.Add("cell", string(cellJson))
				if cfg.Debug {
					log.Println(fmt.Sprintf("%+v", params))
				}
				// Send to document api
				payload := strings.NewReader(params.Encode())
				client := &http.Client{
					Timeout: time.Second * 5,
				}
				req, err := http.NewRequest("POST", cfg.ApiEndpoint+"/document?url_id="+rUrl.Id, payload)
				if err != nil {
					log.Fatalln("Request error: " + err.Error())
				}
				req.Body.Close()
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := client.Do(req)
				if err != nil {
					log.Fatalln("Response error: ", err)
				}
				if resp.StatusCode != 404 {
					respBody, _ := ioutil.ReadAll(resp.Body)
					if resp.StatusCode == 200 {
						log.Println("Callback is successful")
					} else {
						log.Println("Callback is failed, http code is " + resp.Status)
					}
					if cfg.Debug {
						log.Println("Callback return message: " + string(respBody))
					}
				} else {
					log.Println("Not found callback api" + req.URL.String())
				}
				resp.Body.Close()
			}

		}
	} else {
		log.Fatalln("Response status code is " + string(resp.StatusCode))
	}
	log.Println(msg)
}
func main() {
	jsonByte := []byte{}
	ok := false
	switch cfg.DataSource {
	case "api":
		ds := datasource.ApiDataSource{}
		jsonByte, ok = ds.Read()

	default:
		ds := datasource.LocalDataSource{}
		jsonByte, ok = ds.Read()
	}

	if ok {
		resp := response.SuccessResponse{}
		err := json.Unmarshal(jsonByte, &resp)
		if err != nil {
			log.Fatalln("Parse JSON error: ", err)
		}

		log.Println(fmt.Sprintf("%+v", resp))
		urls := resp.Data.Items
		for _, row := range urls {
			wg.Add(1)
			FetchOne(row, &wg)
		}
		wg.Wait()
		log.Println("Done.")
	} else {
		log.Println("Read JSON file error")
	}
}
