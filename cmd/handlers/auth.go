package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func CallGoogleOAuth(ctx echo.Context) error {
	// url := os.Getenv("GOOGLE_OAUTH_URL")
	clientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	// clientSecret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_OAUTH_REDIRECT_URL")

	return ctx.Redirect(http.StatusFound, fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?client_id=%v&redirect_uri=%v&scope=https://www.googleapis.com/auth/userinfo.profile&response_type=code&access_type=offline", clientID, redirectURL))

}

func GoogleOAuthCallback(ctx echo.Context) error {
	query := ctx.QueryParams()
	code := query["code"]

	// Prepare token request
	requestData := url.Values{
		"code":          {code[0]},
		"client_id":     {os.Getenv("CLIENT_ID")},
		"client_secret": {os.Getenv("CLIENT_SECRET")},
		"redirect_uri":  {os.Getenv("REDIRECT_URI")},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequest("POST", "https://www.googleapis.com/oauth2/v4/token", strings.NewReader(requestData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Execute token request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Call user info API with the access token as Bearer token
	userReq, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return err
	}
	userReq.Header.Add("Authorization", fmt.Sprintf("Bearer %v", result["access_token"]))

	userResp, err := client.Do(userReq)
	if err != nil {
		return err
	}
	defer userResp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:    "access_token",
		Value:   result["access_token"].(string),
		Path:    "/",
		Expires: time.Now().Add(365 * 24 * time.Hour),
		Secure:  true,
	}
	//ctx.SetCookie(cookie)
	ctx.Response().Header().Set("Set-Cookie", cookie.String())

	return ctx.Redirect(http.StatusFound, "/")
}

func Callback(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(200, "test")
}
