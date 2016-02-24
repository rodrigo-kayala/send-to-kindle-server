package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/validator.v2"
)

// MOBIGenerator data type
// required fields:
// BaseDir
// DocumentContent
type MOBIGenerator struct {
	DocumentContent string `validate:"min=1"`
	fileName        string
	BaseDir         string `validate:"min=1"`
	article         *goquery.Document
}

// ConvertToMOBI converts the DocumentContent to MOBI
// and returns generated file name
func (gen MOBIGenerator) ConvertToMOBI() (string, error) {

	if err := validator.Validate(gen); err != nil {
		return "", err
	}

	gen.article = gen.buildHTMLDocument()
	gen.cleanUpHTML()
	gen.fileName = gen.extractFilename()
	gen.downloadImages()
	gen.saveHTML()

	kindlegen := os.Getenv("KINDLE_GEN")

	if kindlegen == "" {
		kindlegen = "kindlegen_macosx"
	}

	cmd := exec.Command("./"+kindlegen, gen.BaseDir+"/"+gen.fileName+".html", "-o", gen.fileName+".mobi")
	ret, _ := cmd.Output()
	fmt.Println(string(ret))
	return gen.fileName, nil
}

func (gen MOBIGenerator) cleanUpHTML() {
	gen.article.Find(".sub").Remove()
}

func (gen MOBIGenerator) saveHTML() {
	html, _ := gen.article.Html()
	ioutil.WriteFile(gen.BaseDir+"/"+gen.fileName+".html", []byte(html), 0644)
}

func (gen MOBIGenerator) extractFilename() string {
	title := gen.article.Find(".reader_head").Find("h1").Text()
	re := regexp.MustCompile("[A-Za-z]+")
	return strings.Join(re.FindAllString(title, -1), "")
}

func (gen MOBIGenerator) buildHTMLDocument() *goquery.Document {
	articleText := "<?xml version=\"1.0\" encoding=\"UTF-8\" ?>"
	articleText += "<html><head>"
	articleText += "<meta http-equiv=\"content-type\" content=\"application/xhtml+xml; charset=UTF-8\"></head>"
	articleText += "<body><div>" + gen.DocumentContent + "</div></body></html>"

	article, _ := goquery.NewDocumentFromReader(strings.NewReader(articleText))
	return article
}

func (gen MOBIGenerator) downloadImages() {
	client := http.Client{}
	imageCount := 1

	gen.article.Find("img").Each(func(i int, s *goquery.Selection) {

		out, _ := os.Create(gen.BaseDir + "/" + gen.fileName + strconv.Itoa(imageCount) + ".jpg")
		defer out.Close()

		url := s.AttrOr("src", "")
		s.SetAttr("src", gen.fileName+strconv.Itoa(imageCount)+".jpg")

		reqImg, _ := client.Get(url)
		defer reqImg.Body.Close()
		imageCount++

		io.Copy(out, reqImg.Body)

	})
}
