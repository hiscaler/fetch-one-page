package main

import (
	"io/ioutil"
	"log"
	"response"
	"encoding/json"
	"fmt"
	"strings"
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
				fmt.Println("> " + url)
				project := row.Project
				msg := "> Page Render method: " + project.PageRenderMethod + ", Use Agent: "
				if project.UseAgent {
					msg += "Yes"
				} else {
					msg += "No"
				}
				fmt.Println(msg)
				fmt.Println(strings.Repeat("#", 80))
			}
			fmt.Println("Done.")
		}
	} else {
		log.Println("Read JSON file error")
	}
}
