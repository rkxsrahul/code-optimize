package api

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"git.xenonstack.com/util/continuous-security-backend/config"
	"git.xenonstack.com/util/continuous-security-backend/src/database"
	"git.xenonstack.com/util/continuous-security-backend/src/hubspot"
	"git.xenonstack.com/util/continuous-security-backend/src/method"
	"git.xenonstack.com/util/continuous-security-backend/src/web"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
)

//GitScan is an api handler for handle the github scan related API
func GitScan(c *gin.Context) {
	mapd := make(map[string]interface{})
	var data database.RequestInfo
	if err := c.BindJSON(&data); err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"error":   true,
			"message": "Please pass Git URL",
		})
		return
	}

	if c.Request.Header.Get("Authorization") != "" {
		claims, err := method.ExtractClaims(strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "))
		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{
				"error":   true,
				"message": "Please login again",
			})
			return
		}
		//claim the name from the token
		data.Name = claims["name"].(string)
		//claim the email from the token
		data.Email = claims["email"].(string)
	}

	if data.Email == "" {
		c.JSON(400, gin.H{
			"error":   true,
			"message": "Please pass required information",
		})
		return
	}

	//set the worspace name
	data.Workspace = c.Query("workspace")
	if data.Workspace == "" {
		data.Workspace = method.ProjectNamebyEmail(data.Email)
	}

	//validate the github URL
	giturls := strings.Split(data.URL, "/")
	if len(giturls) < 5 {
		mapd["error"] = true
		mapd["message"] = "Please pass the the valid url"
		c.JSON(400, mapd)
		return
	}

	//Fetch the required information form URL
	projectName := giturls[3]
	repoName := giturls[4]
	repos := strings.Split(repoName, ".")
	repoName = repos[0]

	//check the language the username of the provide github link
	language, username, err := method.FetchLanguage(projectName, repoName)
	if err != nil {
		log.Println(err)
		mapd["error"] = true
		mapd["message"] = err.Error()
		c.JSON(400, mapd)
		return
	}
	//check the language
	if strings.Contains("javascript,python,rust,golang,ruby", strings.ToLower(language.Language)) != true {
		mapd["error"] = true
		mapd["message"] = "Please send Github url of Javascript, Python, Rust, Golang"
		c.JSON(400, mapd)
		return
	}

	var firstname, lastname string
	//manage the userName for the request information data
	if data.Name == "" {
		names := strings.Split(username, " ")
		if username == "" {
			firstname = projectName
			lastname = ""
		} else if len(names) == 1 {
			firstname = username
			lastname = ""
		} else {
			firstname = names[0]
			lastname = strings.Join(names[1:], " ")
		}
		data.Name = firstname + " " + lastname
	}

	//set the user Ip and agent information
	data.IP = c.ClientIP()
	data.Agent = c.Request.UserAgent()

	//store the github information and other required information
	info, err := web.GitScan(data, strings.ToLower(language.Language))
	if err != nil {
		mapd["error"] = true
		mapd["message"] = err.Error()
		c.JSON(400, mapd)
		return
	}

	//hubspot entry for github integration with user details
	err = hubspot.HubspotSubmission(data.Email, firstname, lastname, data.URL, "GitScan")
	msg := ""
	if err != nil {
		msg = err.Error()
	}

	c.JSON(200, gin.H{
		"error":         false,
		"data":          info,
		"info":          data,
		"error_message": msg,
	})
}

//ScanResult is an API handler for handle the API related to the scan result of the website
func GitScanResult(c *gin.Context) {
	chanstream := make(chan interface{})
	go web.FetchGitResult(c.Param("id"), chanstream)
	// c.JSON(200, <-chanstream)
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-chanstream; ok {
			c.SSEvent("message", msg)
			return true
		}
		return false
	})
}

//AccessToken is the structure to get the access token of the user
type AccessToken struct {
	Token    string `json:"token"`
	UserName string `json:"username"`
}

//GitRepos is used to get all public and private repos of the user
func GitRepos(c *gin.Context) {
	mapd := make(map[string]interface{})
	//extracting jwt claims
	claims := jwt.ExtractClaims(c)
	id, ok := claims["id"].(string)
	if !ok {
		c.JSON(500, gin.H{"error": true, "message": "Please login again"})
		return
	}

	payload := []byte(id)
	//configure NATS
	nc := config.NC
	//get access token of the user from auth service using NATS
	res, err := nc.Request("access-token", payload, 1*time.Minute)
	if err != nil {
		log.Println(err)
		mapd["error"] = true
		mapd["message"] = err.Error()
		c.JSON(400, mapd)
		return
	}
	response := AccessToken{}
	err = json.Unmarshal(res.Data, &response)
	if err != nil {
		log.Println(err)
		mapd["error"] = true
		mapd["message"] = err.Error()
		c.JSON(400, mapd)
		return
	}
	//fetch all the repos of the user
	data, code := GetRepos(response)
	c.JSON(code, data)
	return

}

//ReposData is a structure to get repos data
type ReposData struct {
	FullName int    `json:"full_name"`
	Private  bool   `json:"private"`
	HtmlUrl  string `json:"html_url"`
	CloneUrl string `json:"clone_url"`
	Name     string `json:"name"`
}

//GetRepos is used to fetch all the repos using access token of the user
func GetRepos(token AccessToken) (map[string]interface{}, int) {
	mapd := make(map[string]interface{})
	url := "https://api.github.com/user/repos"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.Println(err)
		mapd["error"] = true
		mapd["message"] = err.Error()
		return mapd, 400
	}
	req.Header.Add("Authorization", "token "+token.Token)

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		mapd["error"] = true
		mapd["message"] = err.Error()
		return mapd, 400
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		mapd["error"] = true
		mapd["message"] = err.Error()
		return mapd, 400
	}
	var data []ReposData
	json.Unmarshal(body, &data)
	mapd["repositories_data"] = data
	mapd["repositories_count"] = len(data)
	mapd["error"] = false
	return mapd, 200

}
