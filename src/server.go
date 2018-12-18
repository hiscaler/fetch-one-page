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
)

func Read() ([]byte, bool) {
	content, err := ioutil.ReadFile("data/feed.json")
	if err != nil {
		log.Fatalln(err)
		return nil, false
	}

	return content, true
}

func main() {
	jsonByte, ok := Read()
	if ok {
		resp := response.SuccessResponse{}
		err := json.Unmarshal(jsonByte, &resp)
		if err != nil {
			log.Fatalln("Parse JSON error: ", err)
		}

		log.Println(fmt.Sprintf("%+v", resp))
		urls := resp.Data.Items
		for _, row := range urls {
			urlPath := row.Url
			log.Println("> " + urlPath)
			project := row.Project
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
					// Parse source code
					doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
					if err != nil {
						log.Fatal(err)
					}
					cell := make(map[string]string, len(project.Props))
					for _, prop := range project.Props {
						text := ""
						for _, rule := range prop.Rules {
							if len(text) > 0 {
								break
							}
							if rule.RuleType == "css" {
								// Single
								switch rule.Parser {
								case "text", "raw":
									text = doc.Find(rule.Path).Text()
								case "attr":
									text, _ = doc.Find(rule.Path).Attr(rule.Attr)
								}

								// List
								//doc.Find(rule.Path).Each(func(i int, s *goquery.Selection) {
								//	// For each item found, get the band and title
								//	switch rule.Parser {
								//	case "text":
								//	case "raw":
								//		text = s.Text()
								//		if len(text) > 0 {
								//			break
								//		}
								//		fmt.Println(prop.Name + ":" + text)
								//	}
								//})
							} else if rule.RuleType == "xpath" {
								// todo
							}
						}
						cell[prop.Name] = strings.TrimSpace(text)

					}
					params := url.Values{}
					params.Set("url_id", row.Id)
					cellJson, _ := json.Marshal(cell)
					params.Add("cell", string(cellJson))
					log.Println(fmt.Sprintf("%+v", params))
					// Send to document api
					payload := strings.NewReader(params.Encode())
					client := &http.Client{
						Timeout: time.Second * 5,
					}
					req, err := http.NewRequest("POST", "http://192.168.2.222:8003/index.php/document?url_id="+row.Id, payload)
					if err != nil {
						log.Fatalln("Request error: ", err)
					}
					req.Body.Close()
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					resp, err := client.Do(req)
					if err != nil {
						log.Fatalln("Response error: ", err)
					}
					if resp.StatusCode == 200 {
						log.Println("Success")
						respBody, _ := ioutil.ReadAll(resp.Body)
						log.Println("Post api return message: ", string(respBody))
					} else {
						log.Println("Fail")
					}

					resp.Body.Close()
				}
			} else {
				log.Fatalln("Response status code is " + string(resp.StatusCode))
			}
			log.Println(msg)
		}
		log.Println("Done.")
	} else {
		log.Println("Read JSON file error")
	}
}
