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

var rooturl = "https://www.stringmeteo.com/synop/temp_month.php"

func get_page_url(year int, month int) string {
	return fmt.Sprintf("%v?year=%v&month=%v", rooturl, year, month)

}

type StationTemp struct {
	Name string  `csv:"name"`
	Date string  `csv:"date"`
	Temp float64 `csv:"temp"`
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
		s.Find(".small2").Each(func(i int, x *goquery.Selection) {
			txt := x.Text()
			temp, _ := strconv.ParseFloat(txt, 32)
			stationRecords = append(stationRecords, StationTemp{
				Name: stationName,
				Temp: temp,
				Date: fmt.Sprintf("%v-%v-%v", year, month, day),
			})
			day += 1
		})
		allTemps = append(allTemps, stationRecords...)
		log.Println(fmt.Sprintf("Processed %v %v %v", stationName, year, month))
	})

	c.Visit(get_page_url(year, month))
	return allTemps
}

func main() {
	output_file, err := os.OpenFile("temps.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
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
