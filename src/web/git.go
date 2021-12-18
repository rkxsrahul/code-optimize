package web

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

//AccessToken is the structure to get the access token of the user
type AccessToken struct {
	Token    string `json:"token"`
	UserName string `json:"username"`
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
