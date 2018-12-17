package main

import (
	"io/ioutil"
	"log"
	"response"
	"encoding/json"
	"net/http"
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
		} else {
			urls := resp.Data.Items
			for _, row := range urls {
				url := row.Url
				log.Println("> " + url)
				project := row.Project
				msg := "> Page Render method: " + project.PageRenderMethod + ", Use Agent: "
				if project.UseAgent {
					msg += "Yes"
				} else {
					msg += "No"
				}
				response, err := http.Get(url)
				if err != nil {
					log.Fatalln(err)
				}
				if response.StatusCode == 200 {
					content, err := ioutil.ReadAll(response.Body)
					if err != nil {
						log.Fatalln(err)
					} else {
						respBody := string(content)
						log.Println(respBody)
						// Parse source code
					}
				} else {
					log.Fatalln("Response status code is " + string(response.StatusCode))
				}
				log.Println(msg)
			}
			log.Println("Done.")
		}
	} else {
		log.Println("Read JSON file error")
	}
}
