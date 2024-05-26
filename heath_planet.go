package healthplanet

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
)

type HealthPlanet struct {
	LoginId        string
	LoginPasssword string
	ClientId       string
	ClientSecret   string
	Session        *http.Client
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	return value
}

func createClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}
	return client
}

func NewHealthPlanet(loginId, loginPassword, clientId, clientSecret string) *HealthPlanet {
	return &HealthPlanet{
		LoginId:        loginId,
		LoginPasssword: loginPassword,
		ClientId:       clientId,
		ClientSecret:   clientSecret,
		Session:        createClient(),
	}
}

func (hp *HealthPlanet) getAccessToken() (string, error) {
	authUrl, err := url.Parse("https://www.healthplanet.jp/oauth/auth.do")
	if err != nil {
		log.Fatal(err)
	}
	redirectUrl, err := url.Parse("https://www.healthplanet.jp/success.html")
	if err != nil {
		log.Fatal(err)
	}
	authQuery := authUrl.Query()
	authQuery.Set("redirect_uri", redirectUrl.String())
	authQuery.Set("response_type", "code")
	authQuery.Set("client_id", hp.ClientId)
	authQuery.Set("scope", "innerscan")
	authUrl.RawQuery = authQuery.Encode()

	loginUrl, err := url.Parse("https://www.healthplanet.jp/login_oauth.do")
	if err != nil {
		log.Fatal(err)
	}
	loginQuery := loginUrl.Query()
	loginQuery.Set("loginId", hp.LoginId)
	loginQuery.Set("passwd", hp.LoginPasssword)
	loginQuery.Set("send", "1")
	loginQuery.Set("url", authUrl.String())
	loginUrl.RawQuery = loginQuery.Encode()

	fmt.Println(loginUrl.String())

	resp, err := hp.Session.PostForm(loginUrl.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	redirectedUrl, err := url.Parse(resp.Request.URL.String())
	fmt.Println(redirectedUrl)
	fmt.Println(redirectedUrl.Path)
	if err != nil {
		log.Fatal(err)
	}

	if redirectedUrl.Query().Has("error") {
		if redirectedUrl.Query().Get("error") == "invalid_client" {
			return "", fmt.Errorf("Invalid client")
		} else {
			return "", fmt.Errorf("Failed to login due to client")
		}
	}

	oauthToken, err := getOauthTokenFromHtmlDoc(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	approvalUrl, err := url.Parse("https://www.healthplanet.jp/oauth/approval.do")
	if err != nil {
		log.Fatal(err)
	}
	approvalQuery := approvalUrl.Query()
	approvalQuery.Set("oauth_token", oauthToken)
	approvalQuery.Set("approval", "true")
	approvalUrl.RawQuery = approvalQuery.Encode()

	resp, err = hp.Session.PostForm(approvalUrl.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	redirectedUrl = resp.Request.URL
	accessToken := redirectedUrl.Query().Get("code")
	return accessToken, nil
}

func getOauthTokenFromHtmlDoc(body io.ReadCloser) (string, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return "", err
	}

	var oauthToken string
	doc.Find("input[name=oauth_token]").Each(func(_ int, s *goquery.Selection) {
		oauthToken, _ = s.Attr("value")
	})

	if oauthToken == "" {
		return "", fmt.Errorf("Failed to get oauth token")
	}
	return oauthToken, nil
}


func (hp *HealthPlanet) Run() {
	accessToken, err := hp.getAccessToken()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(accessToken)
}
