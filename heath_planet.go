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
	AccessToken string
}

func createClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}
	return client
}

func NewHealthPlanet(loginId, loginPassword, clientId, clientSecret string) *HealthPlanet {
	session := createClient()
	oauthToken, err := getOauthToken(clientId, loginId, loginPassword, session)
	if err != nil {
		log.Fatal(err)
	}

	authCode, err := getAuthCode(oauthToken, session)
	if err != nil {
		log.Fatal(err)
	}

	accessToken, err := getAccessToken(authCode, clientId, clientSecret, session)
	return &HealthPlanet{accessToken}
}

func getOauthToken(clientId string, loginId string, loginPassword string, session *http.Client) (string, error) {
	authUrl := "https://www.healthplanet.jp/oauth/auth.do"
	redirectUrl := "https://www.healthplanet.jp/success.html"

	authQuery := url.Values{}
	authQuery.Set("redirect_uri", redirectUrl)
	authQuery.Set("response_type", "code")
	authQuery.Set("client_id", clientId)
	authQuery.Set("scope", "innerscan")

	authUrlParsed, err := url.Parse(authUrl)
	authUrlParsed.RawQuery = authQuery.Encode()

	loginUrl := "https://www.healthplanet.jp/login_oauth.do"

	loginQuery := url.Values{}
	loginQuery.Set("loginId", loginId)
	loginQuery.Set("passwd", loginPassword)
	loginQuery.Set("send", "1")
	loginQuery.Set("url", authUrlParsed.String())

	loginUrlParsed, err := url.Parse(loginUrl)
	loginUrlParsed.RawQuery = loginQuery.Encode()

	resp, err := session.PostForm(loginUrlParsed.String(), nil)
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

func getAuthCode(oauthToken string, session *http.Client) (string, error) {
	approvalUrl, err := url.Parse("https://www.healthplanet.jp/oauth/approval.do")
	if err != nil {
		return "", err
	}
	approvalQuery := approvalUrl.Query()
	approvalQuery.Set("oauth_token", oauthToken)
	approvalQuery.Set("approval", "true")
	approvalUrl.RawQuery = approvalQuery.Encode()

	resp, err := session.PostForm(approvalUrl.String(), nil)
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

func getAccessToken(authCode string, clientId string, clientSecret string, session *http.Client) (string, error) {
	accessTokenUrl := "https://www.healthplanet.jp/oauth/token"

	accessTokenQuery := url.Values{}
	accessTokenQuery.Set("client_id", clientId)
	accessTokenQuery.Set("client_secret", clientSecret)
	accessTokenQuery.Set("code", authCode)
	accessTokenQuery.Set("redirect_uri", "https://www.healthplanet.jp/success.html")
	accessTokenQuery.Set("grant_type", "authorization_code")

	resp, err := session.PostForm(accessTokenUrl, accessTokenQuery)
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

type Status struct {
	BirthDate string `json:"birth_date"`
	Data      []struct {
		Date    string `json:"date"`
		KeyData string `json:"keydata"`
		Model   string `json:"model"`
		Tag     string `json:"tag"`
	}
}

type GetStatusRequest struct {
	DateMode    string
	From        string
	To          string
}

func (hp *HealthPlanet) GetInnerscan(request GetStatusRequest) (*Status, error) {
	innerscanUrl := "https://www.healthplanet.jp/status/innerscan.json"

	postBody := url.Values{}
	postBody.Set("access_token", hp.AccessToken)
	postBody.Set("date", request.DateMode)
	postBody.Set("from", request.From)
	postBody.Set("to", request.To)

	resp, err := http.PostForm(innerscanUrl, postBody)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	byteArray, _ := io.ReadAll(resp.Body)
	jsonBytes := []byte(byteArray)
	status := new(Status)

	if err := json.Unmarshal(jsonBytes, status); err != nil {
		return nil, err
	}

	return status, nil
}
