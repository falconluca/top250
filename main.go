package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

type (
	PageInfo struct {
		Page int
		Url  string
	}

	Movie struct {
		Title    string
		Subtitle string
		Other    string
		Desc     string
		Year     string
		Area     string
		Tag      string
		Star     string
		Comment  string
		Quote    string
	}

	MovieCrawler struct {
		targetUrl string
		pages     *[]PageInfo
		movies    *[]Movie
	}
)

const (
	UserAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36`
)

func NewMovieCrawler() *MovieCrawler {
	return &MovieCrawler{
		targetUrl: "https://movie.douban.com/top250",
		pages:     &[]PageInfo{},
		movies:    &[]Movie{},
	}
}

func (c *MovieCrawler) Pages() *[]PageInfo {
	return c.pages
}

func (c *MovieCrawler) printTop250Movies() {
	if len(*c.movies) == 0 {
		pDoc := c.fetchPages()
		c.parsePages(pDoc)
		for _, page := range *c.Pages() {
			mDoc := c.fetchMovies(page)
			c.parseMovies(mDoc)
		}
	}
	for i, movie := range *c.movies {
		log.Printf("No.%d %+v", i+1, movie)
	}
}

func (c *MovieCrawler) fetchPages() *goquery.Document {
	doc, err := handleRequest(http.MethodGet, c.targetUrl, nil)
	if err != nil {
		exitWithErrorf("doc parse failed: %v", err)
	}
	return doc
}

func (c *MovieCrawler) parsePages(doc *goquery.Document) {
	page := PageInfo{
		Page: 1,
		Url:  "",
	}
	newPages := append(*c.pages, page)

	doc.Find("#content > div > div.article > div.paginator > a").
		Each(func(i int, s *goquery.Selection) {
			page, err := strconv.Atoi(s.Text())
			if err != nil {
				log.Fatalf("Atoi failed: %v", err)
			}
			url, exists := s.Attr("href")
			if !exists {
				log.Fatalf("href not exists: %v", exists)
			}

			newPages = append(newPages, PageInfo{
				Page: page,
				Url:  url,
			})
		})

	c.pages = &newPages
}

func (c *MovieCrawler) fetchMovies(page PageInfo) *goquery.Document {
	url := strings.Join([]string{c.targetUrl, page.Url}, "")
	doc, err := handleRequest(http.MethodGet, url, nil)
	if err != nil {
		exitWithErrorf("doc parse failed: %v", err)
	}
	return doc
}

func (c *MovieCrawler) parseMovies(doc *goquery.Document) {
	var moviesOfCurPage []Movie
	doc.Find("#content > div > div.article > ol > li").
		Each(func(i int, s *goquery.Selection) {
			title := s.Find(".hd a span").Eq(0).Text()

			subtitle := s.Find(".hd a span").Eq(1).Text()
			subtitle = strings.TrimLeft(subtitle, "  / ")

			other := s.Find(".hd a span").Eq(2).Text()
			other = strings.TrimLeft(other, "  / ")

			desc := strings.TrimSpace(s.Find(".bd p").Eq(0).Text())
			DescInfo := strings.Split(desc, "\n")
			desc = DescInfo[0]

			movieDesc := strings.Split(DescInfo[1], "/")
			year := strings.TrimSpace(movieDesc[0])
			area := strings.TrimSpace(movieDesc[1])
			tag := strings.TrimSpace(movieDesc[2])

			star := s.Find(".bd .star .rating_num").Text()

			comment := strings.TrimSpace(s.Find(".bd .star span").Eq(3).Text())
			p := regexp.MustCompile("[0-9]")
			comment = strings.Join(p.FindAllString(comment, -1), "")

			quote := s.Find(".quote .inq").Text()

			movie := Movie{
				Title:    title,
				Subtitle: subtitle,
				Other:    other,
				Desc:     desc,
				Year:     year,
				Area:     area,
				Tag:      tag,
				Star:     star,
				Comment:  comment,
				Quote:    quote,
			}
			moviesOfCurPage = append(moviesOfCurPage, movie)
		})

	newMovies := append(*c.movies, moviesOfCurPage...)
	c.movies = &newMovies
}

func handleRequest(httpMethod string, url string, body io.Reader) (*goquery.Document, error) {
	req, err := http.NewRequest(httpMethod, url, body)
	req.Header.Add("User-Agent", UserAgent)
	if err != nil {
		exitWithError(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		exitWithErrorf("do http request failed: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func exitWithErrorf(format string, a ...interface{}) {
	exitWithError(fmt.Errorf(format, a...))
}

func exitWithError(err error) {
	if err != nil {
		color.New(color.FgRed).Fprint(os.Stderr, "Error: ")
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func main() {
	c := NewMovieCrawler()
	c.printTop250Movies()
}
