package main

import (
	"io/ioutil"
	"log"
	"response"
	"encoding/json"
	"fmt"
	"strings"
	"net/http"
	"config"
	"time"
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"net/url"
	"github.com/chromedp/chromedp"
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp/runner"
	"datasource"
)

var cfg config.Config

const (
	PageRenderNormalMethod = "normal"
	PageRenderJsMethod     = "js"
)

const (
	UrlPendingStatus = iota
	UrlWorkingStatus
	UrlFailStatus
	UrlSuccessStatus
	UrlCancelStatus
)

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

var subRegexes map[string]string

// Use chromedp open page, and get page html source
func openPage(url string, pageSourceHtml *string, sleepSeconds time.Duration) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(sleepSeconds * time.Second),
		chromedp.ActionFunc(func(ctx context.Context, h cdp.Executor) error {
			node, err := dom.GetDocument().Do(ctx, h)
			if err != nil {
				return err
			}
			*pageSourceHtml, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx, h)
			return err
		}),
	}
}

func FetchOne(rUrl response.Url) {
	urlPath := rUrl.Url
	log.Println(" >>> " + urlPath)
	project := rUrl.Project
	msg := "Page Render method: " + project.PageRenderMethod + ", Use Agent: "
	if project.UseAgent {
		msg += "Yes"
	} else {
		msg += "No"
	}
	log.Println(msg)
	pageHtmlSource := ""
	if project.PageRenderMethod == PageRenderNormalMethod {
		resp, err := http.Get(urlPath)
		if err != nil {
			log.Fatalln(err)
		}
		if resp.StatusCode == 200 {
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			} else {
				pageHtmlSource = string(respBody)
			}
		} else {
			log.Fatalln("Response status code is " + string(resp.StatusCode))
		}
	} else if project.PageRenderMethod == PageRenderJsMethod {
		var err error
		// create context
		ctxt, cancel := context.WithCancel(context.Background())
		defer cancel()

		// create chrome instance
		//c, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf))
		c, err := chromedp.New(ctxt, chromedp.WithRunnerOptions(
			runner.Flag("headless", cfg.Chromedp.Headless),
			runner.Flag("disable-gpu", cfg.Chromedp.DisableGPU),
		))
		chromedp.WithLog(log.Printf)
		if err != nil {
			log.Fatal(err)
		}

		// run task list
		err = c.Run(ctxt, openPage(urlPath, &pageHtmlSource, 20))
		if err != nil {
			log.Fatal(err)
		}

		// shutdown chrome
		err = c.Shutdown(ctxt)
		if err != nil {
			log.Fatal(err)
		}

		// wait for chrome to finish
		err = c.Wait()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Unkonw page render method: " + project.PageRenderMethod)
	}

	if len(pageHtmlSource) > 0 {
		// Parse page source code
		fmt.Println("pageHtmlSource = ", string(pageHtmlSource))
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(pageHtmlSource))
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
					log.Println("Save document is successful")
					rUrl.Status = UrlWorkingStatus
				} else {
					log.Println("Save document is failed, http code is " + resp.Status)
					rUrl.Status = UrlWorkingStatus
					rUrl.CallbackMessage = "Save document is failed, " + string(respBody)
				}
				if err := rUrl.Callback(cfg.ApiEndpoint); err != nil {
					log.Println("URL callback error" + err.Error())
				}
				if cfg.Debug {
					log.Println("Save document return message: " + string(respBody))
				}
			} else {
				log.Println("Not found save document api" + req.URL.String())
			}
			resp.Body.Close()
		}
		rUrl.Status = UrlSuccessStatus
		if err := rUrl.Callback(cfg.ApiEndpoint); err != nil {
			log.Println("URL callback error" + err.Error())
		}
	} else {
		log.Println("Can't get page HTML source from " + urlPath)
		// Url 内容获取失败回调
		rUrl.Status = UrlFailStatus
		rUrl.CallbackMessage = "Can't get page html source."
		if err := rUrl.Callback(cfg.ApiEndpoint); err != nil {
			log.Println("URL callback error" + err.Error())
		}
	}
}

func producer(url chan response.Url, oneCycle chan bool) {
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
			if row.Status == UrlPendingStatus {
				url <- row
			}
		}
		oneCycle <- true
	} else {
		log.Println("Read JSON file error")
	}
}

func consumer(url response.Url) {
	FetchOne(url)
}

func main() {
	oneCycle := make(chan bool)
	url := make(chan response.Url)
	for {
		go producer(url, oneCycle)
		go func() {
			for {
				select {
				case v := <-url:
					log.Println("Read value is", v)
					consumer(v)
				}
			}
		}()
		<-oneCycle
		log.Println("One cycle is finished. Sleep 10 seconds.")
		time.Sleep(10 * time.Second)
	}

	log.Println("Quit")
}
