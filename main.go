package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/cantara/bragi"
	cloud "github.com/cantara/nerthus/aws"
	"github.com/cantara/nerthus/crypto"
	"github.com/cantara/nerthus/slack"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
}

func main() {
	loadEnv()
	crypto.InitCrypto()

	region := "us-west-2" //"eu-central-1"
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		log.AddError(err).Fatal("While creating aws session")
	}

	var c cloud.AWS
	// Create an EC2 service client.
	c.NewEC2(sess)
	// Create an ELBv2 service client.
	c.NewELB(sess)

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))
	base := r.Group("/nerthus")

	dash := base.Group("/dash")
	{
		dash.StaticFile("/", "./frontend/public/index.html")
		dash.StaticFile("/global.css", "./frontend/public/global.css")
		dash.StaticFile("/favicon.png", "./frontend/public/favicon.png")
		dash.StaticFS("/build", http.Dir("./frontend/public/build"))
	}

	api := base.Group("")
	auth := api.Group("", gin.BasicAuth(gin.Accounts{
		"sindre": "pass",
	}))

	auth.PUT("/server/:application/*server", newAppHandler(&c))
	auth.POST("/key", newKeyHandler())
	auth.POST("/keyCrypt", newKeyCryptHandler())

	/*
		serverName := "devtest-entraos-notification3"
		port := 18840
		uriPath := "notifications"
		elbListenerArn := "arn:aws:elasticloadbalancing:us-west-2:493376950721:listener/app/devtest-events2-lb/a3807cba101b280b/90abaa841820e9b2"
		elbSecurityGroupId := "sg-1325864d"
	*/

	r.Run(":3030")
}

func lateExecuteDeletersWithErrorLogging(object, logMessage string, f func(...string) error, values ...string) func() {
	return func() {
		s := fmt.Sprintf("Cleaning up: %s", object)
		log.Info(s)
		slack.SendStatus(s)
		err := f(values...)
		if err != nil {
			log.AddError(err).Crit(logMessage)
		}
	}
}

func newKeyCryptHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		key, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to read body",
				"error":   err.Error(),
			})
			return
		}
		c.Request.Body.Close()
		pk, err := crypto.Encrypt(string(key))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to Encrypt key",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Encrypted key successfully",
			"data":    pk,
		})
	}
}

type keyBody struct {
	Key string `form:"key" json:"key" xml:"key" binding:"required"`
}

func newKeyHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		var body keyBody
		err := c.ShouldBind(&body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to get requred data from request. Supported formats are: JSON, XML and HTML form",
				"error":   err.Error(),
			})
			return
		}
		pk, err := crypto.Decrypt(body.Data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to decrypt key",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Dekoded key successfully",
			"data":    pk,
		})
	}
}

type application struct {
	Port             int    `form:"port" json:"port" xml:"port" binding:"required"`
	Path             string `form:"path" json:"path" xml:"path" binding:"required"`
	ELBListenerArn   string `form:"elb_listener_arn" json:"elb_listener_arn" xml:"elb_listener_arn" binding:"required"`
	ELBSecurityGroup string `form:"elb_securitygroup_id" json:"elb_securitygroup_id" xml:"elb_securitygroup_id"`
	Key              string `form:"key" json:"key" xml:"key"`
}

func newAppHandler(cld *cloud.AWS) func(*gin.Context) {
	return func(c *gin.Context) {
		appName := c.Param("application")
		server := c.Param("server")[1:]
		if len(server) == 0 {
			server = appName
		}
		if err := cloud.CheckNameLen(server); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Server name is limited by length"),
				"error":   err.Error(),
			})
			return
		}
		var app application
		err := c.ShouldBind(&app)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to get requred data from request. Supported formats are: JSON, XML and HTML form",
				"error":   err.Error(),
			})
			return
		}
		cld.CreateFromScratch(appName, app.Path, app.Port, app.ELBListenerArn, app.ELBSecurityGroup)
	}
}

func NewStack() Stack {
	return Stack{}
}

type Stack struct {
	funcs []func()
}

func (s *Stack) Push(fun func()) {
	s.funcs = append(s.funcs, fun)
}

func (s *Stack) Pop() func() {
	if s.Empty() {
		return nil
	}
	fun := s.Last()
	s.funcs = s.funcs[:s.Len()-1]
	return fun
}

func (s Stack) Len() int {
	return len(s.funcs)
}

func (s Stack) Last() func() {
	if s.Empty() {
		return nil
	}
	return s.funcs[s.Len()-1]
}

func (s Stack) First() func() {
	if s.Empty() {
		return nil
	}
	return s.funcs[0]
}

func (s Stack) Empty() bool {
	return s.Len() == 0
}
