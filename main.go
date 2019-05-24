package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/jszwec/csvutil"
	"log"
	"os"
	"strconv"
)

var url_temps = "https://www.stringmeteo.com/synop/temp_month.php"
var url_prec = "https://www.stringmeteo.com/synop/prec_month.php"
var mode = "prec"

func get_temp_page_url(year int, month int) string {
	return fmt.Sprintf("%v?year=%v&month=%v", url_temps, year, month)
}

func get_prec_page_url(year int, month int) string {
	return fmt.Sprintf("%v?year=%v&month=%v", url_prec, year, month)
}

type StationTemp struct {
	Name    string  `csv:"name"`
	Date    string  `csv:"date"`
	Reading float64 `csv:"reading"`
}

func get_page_data(year int, month int) []StationTemp {
	var allTemps []StationTemp
	c := colly.NewCollector(
		colly.AllowedDomains("www.stringmeteo.com"),
	)
	//Parse the rows
	c.OnHTML("table table tr", func(e *colly.HTMLElement) {
		s := e.DOM
		stationNames := s.Find(".lb").Children()
		if stationNames.Length() == 0 {
			return
		}
		stationName := stationNames.First().First().Text()
		var stationRecords []StationTemp
		day := 1
		//Get the temps
		s.Find("span").Each(func(i int, x *goquery.Selection) {
			txt := x.Text()
			temp, err := strconv.ParseFloat(txt, 32)
			if err != nil {
				return
			}
			if day > 31 {
				return
			}
			stationRecords = append(stationRecords, StationTemp{
				Name:    stationName,
				Reading: temp,
				Date:    fmt.Sprintf("%v-%v-%v", year, month, day),
			})
			day += 1
		})
		allTemps = append(allTemps, stationRecords...)
		log.Println(fmt.Sprintf("Processed %v %v %v", stationName, year, month))
	})
	if mode == "temp" {
		c.Visit(get_temp_page_url(year, month))
	} else if mode == "prec" {
		c.Visit(get_prec_page_url(year, month))
	}

	return allTemps
}

func main() {
	output_fname := ""
	if mode == "temp" {
		output_fname = "temps.csv"
	} else if mode == "prec" {
		output_fname = "prec.csv"
	}

	output_file, err := os.OpenFile(output_fname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}
	startYear := 2018
	endYear := 2019
	endMonth := 5
	startMonth := 1

	crYear := startYear
	crMonth := startMonth

	done := make(chan int)

	var allTemps []StationTemp
	go func() {
		for {
			monthTemps := get_page_data(crYear, crMonth)
			if crYear == endYear && crMonth == endMonth {
				done <- 1
				return
			}
			crMonth += 1
			if crMonth > 12 {
				crMonth = 1
				crYear += 1
			}
			allTemps = append(allTemps, monthTemps...)
		}
	}()

	<-done
	b, err := csvutil.Marshal(allTemps)
	if err != nil {
		log.Println(err)
		return
	}
	output_file.Write(b)
}
