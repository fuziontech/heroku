package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// simulate some private data
var secrets = gin.H{
	"foo":    gin.H{"email": "foo@bar.com", "phone": "123433"},
	"austin": gin.H{"email": "austin@example.com", "phone": "666"},
	"lena":   gin.H{"email": "lena@guapa.com", "phone": "523443"},
}

func main() {
	r := gin.Default()

	// Group using gin.BasicAuth() middleware
	// gin.Accounts is a shortcut for map[string]string
	authorized := r.Group("/admin", gin.BasicAuth(gin.Accounts{
		"foo":    "bar",
		"austin": "1234",
		"lena":   "hello2",
		"manu":   "4321",
	}))

	heroku := r.Group("/heroku", BasicAuthForHeroku())

	heroku.POST("/resources", func(c *gin.Context) {
		buf, _ := c.GetRawData()
		body := string(buf)
		fmt.Printf("%v+", body)
		c.JSON(http.StatusBadRequest, nil)
	})

	authorized.GET("/secrets", func(c *gin.Context) {
		// get user, it was set by the BasicAuth middleware
		user := c.MustGet(gin.AuthUserKey).(string)
		if secret, ok := secrets[user]; ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "secret": secret})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "secret": "NO SECRET :("})
		}
	})

	r.GET("/shit", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"hi": "there"})
	})
	r.Run()
}

type HerokuApiConfig struct {
	ConfigVarsPrefix string
	ConfigVars       []string
	Password         string
	SsoSalt          string
	Regions          []string
	Requires         []string
	Production       map[string]string
	Version          int
}

type HerokuConfig struct {
	Id   string
	Name string
	API  HerokuApiConfig
}

func BasicAuthForHeroku() gin.HandlerFunc {
	realm := "Authorization Required"
	realm = "Basic realm=" + strconv.Quote(realm)
	config := getHerokuConfig()

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		encodedUserPass := strings.Split(authHeader, " ")[1]
		usernamePass := strings.Split(encodedUserPass, ":")
		username := usernamePass[0]
		pass := usernamePass[1]

		if username == config.Id && pass == config.API.Password {
			c.Set(gin.AuthUserKey, username)
		}

		if username != config.Id || pass != config.API.Password {
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func getHerokuConfig() HerokuConfig {
	jsonFile, err := os.Open("addon-manifest.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully grabbed addon-manifest.json")
	defer jsonFile.Close()

	configBytes, _ := ioutil.ReadAll(jsonFile)

	var herokuConfig HerokuConfig
	json.Unmarshal(configBytes, &herokuConfig)

	fmt.Printf("%v+\n", herokuConfig)
	return herokuConfig
}
