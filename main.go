package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/cantara/bragi"
	cloud "github.com/cantara/nerthus/aws"
	"github.com/cantara/nerthus/aws/loadbalancer"
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
	since := time.Now()

	region := os.Getenv("region") //"us-west-2" //"eu-central-1"
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

	dash := base.Group("/") //Might need to be in subdir dash
	{
		dash.StaticFile("/", "./frontend/public/index.html")
		dash.StaticFile("/global.css", "./frontend/public/global.css")
		dash.StaticFile("/favicon.png", "./frontend/public/favicon.png")
		dash.StaticFS("/build", http.Dir("./frontend/public/build"))
	}

	api := base.Group("")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":        "UP",
			"version":       "unknown",
			"ip":            "unknown",
			"running_since": since,
			"now":           time.Now(),
		})
	})

	username := os.Getenv("username")
	password := os.Getenv("password")
	if username == "" || password == "" {
		log.Fatal("Missing user config in env file")
	}

	auth := api.Group("", gin.BasicAuth(gin.Accounts{
		username: password,
	}))
	//auth.PUT("/server/:scope/*server", newServerHandler(&c))
	auth.PUT("/server/:scope/:server", newServerInScopeHandler(&c))
	auth.PUT("/service/:scope/:server", newServiceOnServerHandler(&c))
	auth.POST("/key", newKeyHandler(&c))
	auth.POST("/keyCrypt", newKeyCryptHandler())
	auth.GET("/loadbalancers", newLoadbalancerHandler(&c))

	/*
		serverName := "devtest-entraos-notification3"
		port := 18840
		uriPath := "notifications"
		elbListenerArn := "arn:aws:elasticloadbalancing:us-west-2:493376950721:listener/app/devtest-events2-lb/a3807cba101b280b/90abaa841820e9b2"
		elbSecurityGroupId := "sg-1325864d"
	*/

	r.Run(":" + os.Getenv("port"))
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
		pk, err := crypto.Encrypt(key)
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

func newKeyHandler(cld *cloud.AWS) func(*gin.Context) {
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
		scope, _, k, sg, err := cloud.Decrypt(body.Key, cld)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to decrypt key",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":        "Dekoded key successfully",
			"scope":          scope,
			"key":            k,
			"security_group": sg,
		})
	}
}

func newLoadbalancerHandler(cld *cloud.AWS) func(*gin.Context) {
	return func(c *gin.Context) {
		loadbalancers, err := loadbalancer.GetLoadbalancers(cld.GetELB())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something wend wrong while gettign loadbalancers",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":       "Success",
			"loadbalancers": loadbalancers,
		})
	}
}

type scopeReq struct {
	Key       string            `form:"key" json:"key" xml:"key"`
	Service   cloud.Service     `form:"service" json:"service" xml:"service" binding:"required"`
	ISpesProp map[string]string `form:"instance_specific_propperties" json:"instance_specific_propperties" xml:"instance_specific_propperties"`
}

func newServerInScopeHandler(cld *cloud.AWS) func(*gin.Context) {
	return func(c *gin.Context) {
		scope := c.Param("scope")
		server := c.Param("server") //[1:]
		if err := cloud.CheckNameLen(scope); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Scope name is limited by length"),
				"error":   err.Error(),
			})
			return
		}
		var req scopeReq
		err := c.ShouldBind(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to get requred data from request. Supported formats are: JSON, XML and HTML form",
				"error":   err.Error(),
			})
			return
		}
		log.Println(req.Key)
		if req.Key == "" {
			crypData := cld.CreateNewServerInScope(scope, server, req.Service)
			if crypData == "" {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Something wend wrong while creating server",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Server successfully created",
				"key":     crypData,
			})
			return
		}
		cryptScope, v, k, sg, err := cloud.Decrypt(req.Key, cld)
		if err != nil {
			log.AddError(err).Fatal("While dekrypting cryptdata")
		}
		log.Println("Decrypted key")
		if cryptScope != scope {
			log.Fatal("Scope in cryptodata and provided scope are different")
		}
		crypData := cld.AddNewServerToScope(scope, server, v, k, sg, req.Service)
		if crypData == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something wend wrong while creating server",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Server successfully created",
			"key":     crypData,
		})
	}
}

func newServiceOnServerHandler(cld *cloud.AWS) func(*gin.Context) {
	return func(c *gin.Context) {
		scope := c.Param("scope")
		server := c.Param("server") //[1:]
		if err := cloud.CheckNameLen(scope); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Scope name is limited by length"),
				"error":   err.Error(),
			})
			return
		}
		var req scopeReq
		err := c.ShouldBind(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to get requred data from request. Supported formats are: JSON, XML and HTML form",
				"error":   err.Error(),
			})
			return
		}
		cryptScope, v, k, sg, err := cloud.Decrypt(req.Key, cld)
		if err != nil {
			log.AddError(err).Fatal("While dekrypting cryptdata")
		}
		if cryptScope != scope {
			log.Fatal("Scope in cryptodata and provided scope are different")
		}
		crypData := cld.AddServiceToServer(scope, server, v, k, sg, req.Service)
		if crypData == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something wend wrong while creating server",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Server successfully created",
			"key":     crypData,
		})
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

/*
func newServerHandler(cld *cloud.AWS) func(*gin.Context) {
	return func(c *gin.Context) {
		scope := c.Param("scope")
		server := c.Param("server")[1:]
		if err := cloud.CheckNameLen(server); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Server name is limited by length"),
				"error":   err.Error(),
			})
			return
		}
		var req scopeReq
		err := c.ShouldBind(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to get requred data from request. Supported formats are: JSON, XML and HTML form",
				"error":   err.Error(),
			})
			return
		}
		/*
			cld.CreateNewServerInScope() // With neither key nor server
			cld.AddNewServiceToServer()  // With key and server
			cld.AddNewServerToScope()    // With only key
	}
}
*/
