package healthplanet

import (
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"
	"github.com/joho/godotenv"
	"fmt"
)

type testEnv struct {
	LoginId     string
	LoginPass   string
	ClientId    string
	ClientSecret string
}

func loadTestEnv(t *testing.T) testEnv {
	err := godotenv.Load()
	if err != nil {
		t.Logf("Warning: .envファイルの読み込みに失敗しました: %v", err)
	}
	e := testEnv{
		LoginId:     os.Getenv("HEALTHPLANET_LOGIN_ID"),
		LoginPass:   os.Getenv("HEALTHPLANET_LOGIN_PASSWORD"),
		ClientId:    os.Getenv("HEALTHPLANET_CLIENT_ID"),
		ClientSecret: os.Getenv("HEALTHPLANET_CLIENT_SECRET"),
	}
	if e.LoginId == "" || e.LoginPass == "" || e.ClientId == "" || e.ClientSecret == "" {
		t.Fatal("必要な環境変数が設定されていません")
	}
	return e
}

func TestGetOauthToken(t *testing.T) {
	e := loadTestEnv(t)
	jar, _ := cookiejar.New(nil)
	session := &http.Client{Jar: jar}

	var token string
	t.Run("getOauthToken", func(t *testing.T) {
		var err error
		token, err = getOauthToken(e.ClientId, e.LoginId, e.LoginPass, session)
		if err != nil {
			t.Fatalf("getOauthToken失敗: %v", err)
		}
		if token == "" {
			t.Fatal("トークンが取得できませんでした")
		}
	})

	var authCode string
	t.Run("getAuthCode", func(t *testing.T) {
		var err error
		authCode, err = getAuthCode(token, session)
		if err != nil {
			t.Fatalf("getAuthCode失敗: %v", err)
		}
		if authCode == "" {
			t.Fatal("認可コードが取得できませんでした")
		}
	})

	t.Run("getAccessToken", func(t *testing.T) {
		accessToken, err := getAccessToken(authCode, e.ClientId, e.ClientSecret, session)
		if err != nil {
			t.Fatalf("getAccessToken失敗: %v", err)
		}
		if accessToken == "" {
			t.Fatal("アクセストークンが取得できませんでした")
		}
	})
}
