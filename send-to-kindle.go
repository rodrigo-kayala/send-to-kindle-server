package main

import (
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

const baseDir string = "/tmp"

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "6060"
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(corsMiddleware())

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

func sendMOBIToKindle(fileName string, data *PageData) {
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

func upload(c *gin.Context) {
	var data PageData
	c.BindJSON(&data)

	gen := MOBIGenerator{
		DocumentContent: data.Data,
		BaseDir:         baseDir,
	}

	fileName, err := gen.ConvertToMOBI()
	if err != nil {
		panic(err)
	}

	sendMOBIToKindle(fileName, &data)
}
