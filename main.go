package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)
var searchKeyword string = "html"
var baseURL string = "https://kr.indeed.com/jobs?q=" + url.QueryEscape(searchKeyword)

type extractedJob struct {
	id string
	title string
	companyName string
	companyLocation string
	salary string
}

func writeJobs(jobs []extractedJob) {
	file,err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID","TITLE","COMPANY_NAME","COMPANY_LOC","SALARY"}
	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://kr.indeed.com/viewjob?jk="+job.id + " ",job.title,job.companyName,job.companyLocation,job.salary}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func getPage(page int, URL string,c chan<- []extractedJob)   {
	cc := make(chan extractedJob)
	jobs := []extractedJob{}
	pageURL:=URL + "&start=" +  strconv.Itoa(page*50)
	fmt.Println(pageURL)
	res,err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()
	Doc,err := goquery.NewDocumentFromReader(res.Body)
	searchCards := Doc.Find(".mosaic-provider-jobcards>a")
	searchCards.Each(func(i int, s *goquery.Selection){
		go extracteJob(s,cc)
	})
	for i := 0; i < searchCards.Length(); i++ {
		job := <-cc
		jobs = append(jobs,job)
	}
	c <- jobs
}

func getPages(url string) int {
	pages := 0
	res,err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()
	Doc,err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	Doc.Find(".pagination").Find(".pagination-list").Each(func(i int, s *goquery.Selection){
		pages = s.Find("a").Length()
	})
	return pages
}
func checkErr(err error)  {
	if err != nil {
		log.Fatalln(err)
	}
}
func checkCode(res *http.Response)  {
	if res.StatusCode != 200 {
		log.Fatalln("Request Failed! Status:",res.StatusCode)
	}
}

func extracteJob(s *goquery.Selection,c chan<- extractedJob) {
	id,_ := s.Attr("data-jk")
	title := s.Find("div>div>div>div>.jobCard_mainContent>tbody>tr>.resultContent>.singleLineTitle>h2").Text()
	companyName := s.Find("div>div>div>div>.jobCard_mainContent>tbody>tr>.resultContent>.company_location>pre>.companyName").Text()
	companyLocation := s.Find("div>div>div>div>.jobCard_mainContent>tbody>tr>.resultContent>.company_location>pre>.companyLocation").Text()
	salary := s.Find("div>div>div>div>.jobCard_mainContent>tbody>tr>.resultContent>.salary-snippet-container>span").Text()
	c <- extractedJob{id: id,title: title,companyName: companyName,companyLocation: companyLocation,salary: salary}
}

func main() {
	c := make(chan []extractedJob)
	jobs := []extractedJob{}
	totalPages:=getPages(baseURL)
	for i := 0; i < totalPages; i++ {
		go getPage(i,baseURL,c)
	}
	for i := 0; i < totalPages; i++ {
		jobs = append(jobs,<-c...)
	}
	writeJobs(jobs)
	fmt.Println("Done extracted! :3 /",strconv.Itoa(len(jobs))+"(amounts)")
}

