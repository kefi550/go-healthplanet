package healthplanet

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type HealthPlanet struct {
	LoginId        string
	LoginPasssword string
	ClientId       string
	ClientSecret   string
	Session        *http.Client
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

func (hp *HealthPlanet) getOauthToken() (string, error) {
	authUrl := "https://www.healthplanet.jp/oauth/auth.do"
	redirectUrl := "https://www.healthplanet.jp/success.html"

	authQuery := url.Values{}
	authQuery.Set("redirect_uri", redirectUrl)
	authQuery.Set("response_type", "code")
	authQuery.Set("client_id", hp.ClientId)
	authQuery.Set("scope", "innerscan")

	authUrlParsed, err := url.Parse(authUrl)
	authUrlParsed.RawQuery = authQuery.Encode()

	loginUrl := "https://www.healthplanet.jp/login_oauth.do"

	loginQuery := url.Values{}
	loginQuery.Set("loginId", hp.LoginId)
	loginQuery.Set("passwd", hp.LoginPasssword)
	loginQuery.Set("send", "1")
	loginQuery.Set("url", authUrlParsed.String())

	loginUrlParsed, err := url.Parse(loginUrl)
	loginUrlParsed.RawQuery = loginQuery.Encode()

	resp, err := hp.Session.PostForm(loginUrlParsed.String(), nil)
	if err != nil {
		return "", err
	}

	redirectedUrl, err := url.Parse(resp.Request.URL.String())
	if err != nil {
		return "", err
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
		return "", err
	}

	return oauthToken, nil
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

func (hp *HealthPlanet) getAuthCode(oauthToken string) (string, error) {
	approvalUrl, err := url.Parse("https://www.healthplanet.jp/oauth/approval.do")
	if err != nil {
		return "", err
	}
	approvalQuery := approvalUrl.Query()
	approvalQuery.Set("oauth_token", oauthToken)
	approvalQuery.Set("approval", "true")
	approvalUrl.RawQuery = approvalQuery.Encode()

	resp, err := hp.Session.PostForm(approvalUrl.String(), nil)
	if err != nil {
		return "", err
	}

	redirectedUrl := resp.Request.URL
	authCode := redirectedUrl.Query().Get("code")
	if authCode == "" {
		return "", fmt.Errorf("Failed to get access token")
	}

	return authCode, nil
}

type AccessToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (hp *HealthPlanet) getAccessToken(authCode string) (string, error) {
	accessTokenUrl := "https://www.healthplanet.jp/oauth/token"

	accessTokenQuery := url.Values{}
	accessTokenQuery.Set("client_id", hp.ClientId)
	accessTokenQuery.Set("client_secret", hp.ClientSecret)
	accessTokenQuery.Set("code", authCode)
	accessTokenQuery.Set("redirect_uri", "https://www.healthplanet.jp/success.html")
	accessTokenQuery.Set("grant_type", "authorization_code")

	resp, err := hp.Session.PostForm(accessTokenUrl, accessTokenQuery)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	byteArray, _ := io.ReadAll(resp.Body)
	jsonBytes := []byte(byteArray)
	data := new(AccessToken)

	if err := json.Unmarshal(jsonBytes, data); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return "", err
	}
	accessToken := data.AccessToken
	return accessToken, nil
}

func (hp *HealthPlanet) Run() {
	oauthToken, err := hp.getOauthToken()
	if err != nil {
		log.Fatal(err)
	}

	authCode, err := hp.getAuthCode(oauthToken)
	if err != nil {
		log.Fatal(err)
	}

	accessToken, err := hp.getAccessToken(authCode)
	fmt.Println(accessToken)
}
