package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cantara/nerthus/aws/metadata"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	log "github.com/cantara/bragi"
	cloud "github.com/cantara/nerthus/aws"
	"github.com/cantara/nerthus/aws/loadbalancer"
	"github.com/cantara/nerthus/crypto"
	"github.com/cantara/nerthus/slack"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var Version string

var BuildTime string

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func main() {
	loadEnv()
	slack.NewClient(os.Getenv("slack_token"), os.Getenv("slack_channel_secret"), os.Getenv("slack_channel_status"), os.Getenv("slack_channel_commands"))
	crypto.InitCrypto()
	since := time.Now()

	//region := os.Getenv("region") //"us-west-2" //"eu-central-1"
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	var opts []func(*config.LoadOptions) error
	if os.Getenv("aws.profile") != "" {
		opts = append(opts, config.WithSharedConfigProfile(os.Getenv("aws.profile")))
	} else {
		opts = append(opts, config.WithDefaultRegion(os.Getenv("region")))
	}
	sess, err := config.LoadDefaultConfig(context.TODO(), opts...,
	)
	if err != nil {
		log.AddError(err).Fatal("While creating aws session")
	}

	var c cloud.AWS
	// Create an EC2 service client.
	c.NewEC2(sess)
	// Create an ELBv2 service client.
	c.NewELB(sess)
	// Create an rds service client.
	c.NewRDS(sess)

	ids, err := metadata.GetAllServersWithMetadataV1IDs(c.GetEC2())
	if err != nil {
		log.AddError(err).Fatal("while getting all server ids")
		return
	}
	log.Println(ids)
	for _, id := range ids {
		err = metadata.SetMetadataV2(id, c.GetEC2())
		if err != nil {
			log.AddError(err).Fatal("while setting server ", id, " metadata to version 2, ABORTING")
			return
		}
	}

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))
	basePath := ""
	if os.Getenv("run_as_base") != "true" {
		basePath = "/nerthus"
	}
	base := r.Group(basePath)

	dash := base.Group("")
	{
		dash.StaticFile("", "./frontend"+os.Getenv("frontend_path")+"/index.html")
		dash.StaticFile("/global.css", "./frontend"+os.Getenv("frontend_path")+"/global.css")
		dash.StaticFile("/favicon.png", "./frontend"+os.Getenv("frontend_path")+"/favicon.png")
		dash.StaticFS("/build", http.Dir("./frontend"+os.Getenv("frontend_path")+"/build"))
	}

	outboudIp := GetOutboundIP()
	api := base.Group("")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":        "UP",
			"version":       Version,
			"build_time":    BuildTime,
			"ip":            outboudIp.String(),
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
	auth.PUT("/scope/:scope", newScopeHandler(&c))
	auth.PUT("/server/:scope/:server", newServerInScopeHandler(&c))
	auth.PUT("/service/:scope/:server/:service", newServiceOnServerHandler(&c))
	auth.PUT("/database/:scope/:artifactId", newDatabaseInScopeHandler(&c))
	auth.POST("/key", newKeyHandler(&c))
	auth.POST("/keyCrypt", newKeyCryptHandler())
	auth.GET("/loadbalancers", newLoadbalancerHandler(&c))
	auth.GET("/dns/:scope/:server", dnsHandler(&c))

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
		scope, _, k, sg, _, err := cloud.Decrypt(body.Key, cld)
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

func dnsHandler(cld *cloud.AWS) func(*gin.Context) {
	return func(c *gin.Context) {
		scope := c.Param("scope")
		server := c.Param("server")
		publicDNS, err := cloud.GetPublicDNS(server, scope, cld)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to find server",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":    "Public DNS found",
			"public_dns": publicDNS,
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

func newScopeHandler(cld *cloud.AWS) func(*gin.Context) {
	return func(c *gin.Context) {
		scope := c.Param("scope")
		if err := cloud.CheckNameLen(scope); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Scope name is limited by length"),
				"error":   err.Error(),
			})
			return
		}
		go slack.SendCommand(fmt.Sprintf("scope/%s", scope), "")
		crypData := cld.CreateScope(scope)
		if crypData == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something went wrong while creating scope",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Server successfully created",
			"key":     crypData,
		})
		return
	}
}

type serverReq struct {
	Key string `form:"key" json:"key" xml:"key"`
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
		var req serverReq
		err := c.ShouldBind(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to get requred data from request. Supported formats are: JSON, XML and HTML form",
				"error":   err.Error(),
			})
			return
		}
		if req.Key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Body with key and service is required",
			})
			return
		}
		body, _ := json.Marshal(req)
		go slack.SendCommand(fmt.Sprintf("server/%s/%s", scope, server), string(body))
		cryptScope, v, k, sg, ts, err := cloud.Decrypt(req.Key, cld)
		if err != nil {
			log.AddError(err).Fatal("While dekrypting cryptdata")
		}
		log.Println("Decrypted key")
		if cryptScope != scope {
			log.Fatal("Scope in cryptodata and provided scope are different")
		}
		crypData := cld.AddServerToScope(scope, server, v, k, sg, ts)
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

func newDatabaseInScopeHandler(cld *cloud.AWS) func(*gin.Context) {
	return func(c *gin.Context) {
		scope := c.Param("scope")
		artifactId := c.Param("artifactId") //[1:]
		if err := cloud.CheckNameLen(scope); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Scope name is limited by length"),
				"error":   err.Error(),
			})
			return
		}
		var req serverReq
		err := c.ShouldBind(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to get requred data from request. Supported formats are: JSON, XML and HTML form",
				"error":   err.Error(),
			})
			return
		}
		if req.Key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Body with key and service is required",
			})
			return
		}
		body, _ := json.Marshal(req)
		go slack.SendCommand(fmt.Sprintf("database/%s/%s", scope, artifactId), string(body))
		cryptScope, v, _, sg, slackId, err := cloud.Decrypt(req.Key, cld)
		if err != nil {
			log.AddError(err).Fatal("While dekrypting cryptdata")
		}
		log.Println("Decrypted key")
		if cryptScope != scope {
			log.Fatal("Scope in cryptodata and provided scope are different")
		}
		endpoint := cld.CreateDatabase(scope, artifactId, v, sg, slackId)
		if endpoint == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Something wend wrong while creating database",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":  "Database successfully created",
			"endpoint": endpoint,
		})
	}
}

type serviceReq struct {
	Key       string            `form:"key" json:"key" xml:"key"`
	Service   cloud.Service     `form:"service" json:"service" xml:"service" binding:"required"`
	ISpesProp map[string]string `form:"instance_specific_propperties" json:"instance_specific_propperties" xml:"instance_specific_propperties"`
}

func newServiceOnServerHandler(cld *cloud.AWS) func(*gin.Context) {
	return func(c *gin.Context) {
		scope := c.Param("scope")
		server := c.Param("server")
		service := c.Param("service")
		if err := cloud.CheckNameLen(scope); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Scope name is limited by length"),
				"error":   err.Error(),
			})
			return
		}
		var req serviceReq
		err := c.ShouldBind(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to get requred data from request. Supported formats are: JSON, XML and HTML form",
				"error":   err.Error(),
			})
			return
		}
		body, _ := json.Marshal(req)
		go slack.SendCommand(fmt.Sprintf("service/%s/%s/%s", scope, server, service), string(body))
		cryptScope, v, k, sg, ts, err := cloud.Decrypt(req.Key, cld)
		if err != nil {
			log.AddError(err).Fatal("While dekrypting cryptdata")
		}
		if cryptScope != scope {
			log.Fatal("Scope in cryptodata and provided scope are different")
		}
		if req.Service.ArtifactId != service {
			log.Fatal("Artifact id and provided service does not match")
		}
		crypData := cld.AddServiceToServer(scope, server, v, k, sg, ts, req.Service)
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
