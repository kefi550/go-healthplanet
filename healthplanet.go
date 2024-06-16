package healthplanet

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

var (
	authUrl        = "https://www.healthplanet.jp/oauth/auth.do"
	redirectUrl    = "https://www.healthplanet.jp/success.html"
	loginUrl       = "https://www.healthplanet.jp/login_oauth.do"
	approvalUrl    = "https://www.healthplanet.jp/oauth/approval.do"
	accessTokenUrl = "https://www.healthplanet.jp/oauth/token"
	innerscanUrl   = "https://www.healthplanet.jp/status/innerscan.json"
)

const (
	Weight = 6021 + iota
	BodyFat
)

var tagMap = map[int64]string{
	Weight:  "Weight",
	BodyFat: "BodyFat",
}

const (
	DateMode_RegisteredDate = "0"
	DateMode_MeasuredDate   = "1"
)

type Client struct {
	HTTPClient  *http.Client
	accessToken string
}

func NewClient(loginId string, loginPassword string, clientId string, clientSecret string) *Client {
	jar, _ := cookiejar.New(nil)
	session := &http.Client{
		Jar: jar,
	}
	oauthToken, err := getOauthToken(clientId, loginId, loginPassword, session)
	if err != nil {
		log.Fatal(err)
	}

	authCode, err := getAuthCode(oauthToken, session)
	if err != nil {
		log.Fatal(err)
	}

	accessToken, err := getAccessToken(authCode, clientId, clientSecret, session)
	if err != nil {
		log.Fatal(err)
	}
	return &Client{
		accessToken: accessToken,
		HTTPClient:  http.DefaultClient,
	}
}

func getOauthToken(clientId string, loginId string, loginPassword string, session *http.Client) (string, error) {
	authQuery := url.Values{}
	authQuery.Set("redirect_uri", redirectUrl)
	authQuery.Set("response_type", "code")
	authQuery.Set("client_id", clientId)
	authQuery.Set("scope", "innerscan")

	authUrlParsed, err := url.Parse(authUrl)
	authUrlParsed.RawQuery = authQuery.Encode()

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
	approvalUrl, err := url.Parse(approvalUrl)
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
	accessTokenQuery := url.Values{}
	accessTokenQuery.Set("client_id", clientId)
	accessTokenQuery.Set("client_secret", clientSecret)
	accessTokenQuery.Set("code", authCode)
	accessTokenQuery.Set("redirect_uri", redirectUrl)
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

func (c *Client) prepRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("access_token", c.accessToken)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, err
}

type GetStatusRequest struct {
	DateMode string
	From     string
	To       string
	Tag      int
}

func (c *Client) GetInnerscan(r GetStatusRequest) (*Status, error) {
	req, err := c.prepRequest(innerscanUrl)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("date_mode", r.DateMode)
	q.Add("from", r.From)
	q.Add("to", r.To)
	q.Add("tag", fmt.Sprintf("%d", r.Tag))
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Status
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetTagValue(tagKey string) (string, error) {
	tagKeyInt, err := strconv.ParseInt(tagKey, 10, 64)
	if err != nil {
		return "", err
	}
	value := tagMap[tagKeyInt]
	if value == "" {
		return "", fmt.Errorf("Failed to get tag value")
	}
	return value, nil
}
