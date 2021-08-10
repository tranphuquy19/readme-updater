package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-co-op/gocron"
)

type GithubCredential struct {
	DefaultMessage string `toml:"default_message"`
	Username       string `toml:"username"`
	Email          string `toml:"email"`
	Name           string `toml:"name"`
	Repo           string `toml:"repo"`
	FilePath       string `toml:"file_path"`
	Token          string `toml:"token"`
	CronExpression string `toml:"cron_expression"`
}

type BlobContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		HTML string `json:"html"`
	} `json:"_links"`
}

type ReqUpdateReadmeBody struct {
	Message   string    `json:"message"`
	Content   string    `json:"content"`
	Sha       string    `json:"sha"`
	Committer Committer `json:"committer"`
	Author    Author    `json:"author"`
}
type Committer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func getCredentials() GithubCredential {
	var githubCredential GithubCredential
	if _, err := toml.DecodeFile(".credentials", &githubCredential); err != nil {
		panic(err)
	}
	return githubCredential
}

func getBlobContent(githubCredential GithubCredential) BlobContent {
	client := http.Client{}
	reqUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", githubCredential.Username, githubCredential.Repo, githubCredential.FilePath)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		panic(err)
	}
	req.Header = http.Header{
		"Accept":        []string{"application/vnd.github.v3+json"},
		"Authorization": []string{"token " + githubCredential.Token},
		"Content-Type":  []string{"application/json"},
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var blobObj BlobContent
	err = json.Unmarshal(body, &blobObj)
	if err != nil {
		panic(err)
	}

	return blobObj
}

func getWeather() string {
	var weatherStr = ""
	res, err := http.Get("https://wttr.in/Danang?format=v2")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("pre").Each(func(i int, s *goquery.Selection) {
		result := s.Text()
		weatherStr = "<pre>" + result + "</pre>"
		fmt.Printf("Review %d: %s\n", i, result)
	})
	return weatherStr
}

func updateNewReadme(githubCredential GithubCredential, blobObj BlobContent) {
	readmeTmplStr, err := ioutil.ReadFile("README.md.tmpl")
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("readme").Parse(string(readmeTmplStr))
	if err != nil {
		panic(err)
	}

	data := struct {
		Weather string
	}{
		Weather: getWeather(),
	}

	var tpl bytes.Buffer
	if err = tmpl.Execute(&tpl, data); err != nil {
		panic(err)
	}

	readmeExecuted := tpl.String()
	contentBase64Str := base64.StdEncoding.EncodeToString([]byte(readmeExecuted))

	reqBodyObj := ReqUpdateReadmeBody{
		Message: githubCredential.DefaultMessage,
		Content: contentBase64Str,
		Sha:     blobObj.Sha,
		Committer: Committer{
			Email: githubCredential.Email,
			Name:  githubCredential.Name,
		},
		Author: Author{
			Email: githubCredential.Email,
			Name:  githubCredential.Name,
		},
	}

	reqBodyJson, err := json.Marshal(reqBodyObj)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(reqBodyJson))

	client := http.Client{}

	reqUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", githubCredential.Username, githubCredential.Repo, githubCredential.FilePath)
	req, err := http.NewRequest("PUT", reqUrl, bytes.NewBuffer(reqBodyJson))
	if err != nil {
		panic(err)
	}

	req.Header = http.Header{
		"Accept":        []string{"application/vnd.github.v3+json"},
		"Authorization": []string{"token " + githubCredential.Token},
		"Content-Type":  []string{"application/json"},
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	resJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(resJson))
	getWeather()
}

func Run(githubCredential GithubCredential) {
	blobObj := getBlobContent(githubCredential)
	updateNewReadme(githubCredential, blobObj)
}

func main() {
	var scheduler = gocron.NewScheduler(time.UTC)
	githubCredential := getCredentials()
	scheduler.Cron(githubCredential.CronExpression).Do(func() { Run(githubCredential) })
	scheduler.StartBlocking()
}
