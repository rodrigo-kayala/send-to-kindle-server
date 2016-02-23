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
	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "6060"
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.POST("/", upload)
	router.Run(":" + port)
}

// PageData lalsdkflsdkf
type PageData struct {
	Data           string `json:"data"`
	KindleAddress  string `json:"kindleEmail"`
	SMTPServer     string `json:"smtpServer"`
	SMTPPort       string `json:"smtpPort"`
	SenderAddress  string `json:"senderAddress"`
	SenderUsername string `json:"senderUsername"`
	SenderPassword string `json:"senderPassword"`
}

func upload(c *gin.Context) {
	var data PageData

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	c.BindJSON(&data)

	baseDir := "/tmp"

	client := http.Client{}

	articleText := "<?xml version=\"1.0\" encoding=\"UTF-8\" ?>"
	articleText += "<html><head>"
	articleText += "<meta http-equiv=\"content-type\" content=\"application/xhtml+xml; charset=UTF-8\"></head>"
	articleText += "<body><div>" + data.Data + "</div></body></html>"

	article, _ := goquery.NewDocumentFromReader(strings.NewReader(articleText))

	fileCount := 1
	article.Find(".sub").Remove()
	title := article.Find(".reader_head").Find("h1").Text()
	_ = "breakpoint"

	re := regexp.MustCompile("[A-Za-z]+")
	fileName := strings.Join(re.FindAllString(title, -1), "")

	article.Find("img").Each(func(i int, s *goquery.Selection) {

		out, _ := os.Create(baseDir + "/" + fileName + strconv.Itoa(fileCount) + ".jpg")
		defer out.Close()

		url := s.AttrOr("src", "")
		s.SetAttr("src", fileName+strconv.Itoa(fileCount)+".jpg")

		reqImg, _ := client.Get(url)
		defer reqImg.Body.Close()
		fileCount++

		io.Copy(out, reqImg.Body)

	})

	html, _ := article.Html()

	ioutil.WriteFile(baseDir+"/"+fileName+".html", []byte(html), 0644)

	kindlegen := os.Getenv("KINDLE_GEN")

	if kindlegen == "" {
		kindlegen = "kindlegen_macosx"
	}

	cmd := exec.Command("./"+kindlegen, baseDir+"/"+fileName+".html", "-o", fileName+".mobi")
	ret, execerr := cmd.Output()
	fmt.Println(string(ret))
	if execerr != nil {
		panic(execerr)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", data.SenderAddress)
	m.SetHeader("To", data.KindleAddress)
	m.SetHeader("Subject", "kindle-sync - "+fileName+".mobi")
	m.SetBody("text/html", baseDir+"/"+fileName+".mobi")
	m.Attach(baseDir + "/" + fileName + ".mobi")
	port, _ := strconv.Atoi(data.SMTPPort)
	d := gomail.NewPlainDialer(data.SMTPServer, port, data.SenderUsername, data.SenderPassword)
	err := d.DialAndSend(m)

	if err != nil {
		panic(err)
	}
}
