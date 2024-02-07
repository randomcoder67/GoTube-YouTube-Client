package network

import (
	"bytes"
	"encoding/json"
	"golang.org/x/net/html"
	"strings"
)

// This file contains functions to parse HTML into usable JSON

// Prettify json for saving to file - move to config package as it's for logging only
func PrettifyString(str string) []byte {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "	"); err != nil {
		panic(err)
	}
	return prettyJSON.Bytes()
}

// Function to extract single JSON section from HTML
func ExtractJSON(inputHTML string, playlist bool) string {
	var numAfter = 7
	if playlist {
		numAfter = 64
	}
	// Create string to hold extracted JSON
	var finalJSONString string = ""
	// Parse HTML
	doc, err := html.Parse(strings.NewReader(inputHTML))
	if err != nil {
		panic(err)
	}
	//err = os.WriteFile("output.html", []byte(inputHTML), 0666)
	// Function to parse HTML and extract JSON
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Data == "script" {
			if len(n.Attr) > 1 {
				if strings.Contains(n.Attr[1].Val, "desktop_polymer") {
					b := n
					// The required JSON is contained in script tag 7 after the matched one
					for i := 0; i < numAfter; i++ {
						b = b.NextSibling
						if b.FirstChild != nil {
							if strings.Contains(b.FirstChild.Data, "var ytInitialData = ") {
								break
							}
						}
					}
					// Format the string properly, extracting only the JSON
					finalJSONString = strings.ReplaceAll(strings.ReplaceAll(b.FirstChild.Data, "var ytInitialData = ", ""), ";", "")
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	//err = os.WriteFile("raw.json", PrettyString(finalJSONString), 0666)
	// Unmarshal the JSON
	return finalJSONString
}

// Function to extract multiple JSON sections from HTML
func ExtractJSONVideoPage(inputHTML string) (string, string) {
	var numAfter = 8
	// Create string to hold extracted JSON
	var finalJSONString string = ""
	var secondJSONString string = ""
	// Parse HTML
	doc, err := html.Parse(strings.NewReader(inputHTML))
	if err != nil {
		panic(err)
	}
	//err = os.WriteFile("output.html", []byte(inputHTML), 0666)
	// Function to parse HTML and extract JSON
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Data == "script" {
			if len(n.Attr) > 1 {
				if strings.Contains(n.Attr[1].Val, "desktop_polymer") {
					b := n
					// The required JSON is contained in script tag 7 after the matched one
					for i := 0; i < numAfter; i++ {
						b = b.NextSibling
					}
					// Format the string properly, extracting only the JSON
					finalJSONString = strings.ReplaceAll(strings.ReplaceAll(b.FirstChild.Data, "var ytInitialData = ", ""), ";", "")
				}
			}
		} else if n.Data == "body" {
			secondJSONString = strings.ReplaceAll(strings.ReplaceAll(n.FirstChild.NextSibling.FirstChild.Data, "var ytInitialPlayerResponse = ", ""), ";", "")
			var lastCurlyBracket int = strings.LastIndex(secondJSONString, "}")
			secondJSONString = secondJSONString[:lastCurlyBracket+1]
			//fmt.Println(secondJSONString)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	//err = os.WriteFile("initialData.json", PrettyString(finalJSONString), 0666)
	//err = os.WriteFile("playerResponse.json", PrettyString(secondJSONString), 0666)
	// Unmarshal the JSON
	return finalJSONString, secondJSONString
}
