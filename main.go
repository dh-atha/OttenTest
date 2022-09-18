package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	output := FetchHtml("https://gist.githubusercontent.com/nubors/eecf5b8dc838d4e6cc9de9f7b5db236f/raw/d34e1823906d3ab36ccc2e687fcafedf3eacfac9/jne-awb.html")
	jsonData, err := json.MarshalIndent(output, "", "    ")
	errHandler(err)

	fmt.Println(string(jsonData))
}

var bulan = map[string]string{
	"01": "Januari",
	"02": "Februari",
	"03": "Maret",
	"04": "April",
	"05": "Mei",
	"06": "Juni",
	"07": "Juli",
	"08": "Agustus",
	"09": "September",
	"10": "Oktober",
	"11": "November",
	"12": "Desember",
}

func FetchHtml(url string) (output Output) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	errHandler(err)

	res, err := http.DefaultClient.Do(req)
	errHandler(err)
	defer res.Body.Close()

	if res.StatusCode > 400 {
		output.Status.Code = res.Status
		output.Status.Message = "Failed fetching html"
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	errHandler(err)
	main := doc.Find("div.main-content")
	table := main.Find("tbody")
	var resData Data
	table.Each(func(i int, s *goquery.Selection) {
		data := s.Find("tr")
		if data.Size() == 9 {
			data.Each(func(i int, s *goquery.Selection) {
				var history History
				tableData := s.Find("td")
				tableData.Each(func(i int, s *goquery.Selection) {
					if i == 0 {
						time := s.Text()
						for i, value := range strings.Split(time, " ") {
							if i == 0 {
								date := strings.Split(value, "-")
								for j, k := len(date)-1, 0; j >= 0; j-- {
									history.CreatedAt += date[j]
									if k == 1 {
										history.Formatted.CreatedAt += bulan[date[k]]
									} else {
										history.Formatted.CreatedAt += date[k]
									}
									if j != 0 {
										history.CreatedAt += "-"
										history.Formatted.CreatedAt += " "
									} else {
										history.Formatted.CreatedAt += ", "
									}
									k++
								}
							} else {
								history.CreatedAt += "T" + value + ":00+07:00"
								history.Formatted.CreatedAt += value + " WIB"
							}
						}
					} else {
						history.Description = s.Text()
						if strings.Contains(s.Text(), "DELIVERED TO") {
							splitString := strings.Split(s.Text(), "[")[1]
							splitString = strings.Split(splitString, "|")[0]
							resData.ReceivedBy = splitString
						}
					}
				})
				resData.Histories = append(resData.Histories, history)
			})
		}
	})

	// Reverse slice
	for i, j := 0, len(resData.Histories)-1; i < j; i, j = i+1, j-1 {
		resData.Histories[i], resData.Histories[j] = resData.Histories[j], resData.Histories[i]
	}

	status := Status{Code: "060101", Message: "Delivery tracking detail fetched successfully"}
	output = Output{
		Status: status,
		Data:   resData,
	}
	return
}

func errHandler(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

type History struct {
	Description string    `json:"description"`
	CreatedAt   string    `json:"created_at"`
	Formatted   Formatted `json:"formatted"`
}

type Formatted struct {
	CreatedAt string `json:"created_at"`
}

type Data struct {
	ReceivedBy string    `json:"receivedBy"`
	Histories  []History `json:"histories"`
}

type Status struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Output struct {
	Status Status `json:"status"`
	Data   Data   `json:"data"`
}
