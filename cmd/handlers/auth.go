package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/girirock/task-planner/cmd/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func login(isGoogle bool, accessToken string) error {
	if isGoogle {
		//Call user info API with the access token as Bearer token
		userReq, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
		if err != nil {
			return err
		}
		userReq.Header.Add("Authorization", fmt.Sprintf("Bearer %v", accessToken))
		client := &http.Client{}
		userResp, err := client.Do(userReq)
		if err != nil {
			return err
		}
		defer userResp.Body.Close()

		// bodyBytes, err := io.ReadAll(userResp.Body)
		// if err != nil {
		// 	return err
		// }

		var userInfo map[string]interface{}
		if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
			return err
		}
	}
	return nil
}

func generateAccessToken(sub string, picture string, name string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     sub,
		"picture": picture,
		"name":    name,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenstring, err := token.SignedString([]byte("ZG6wkVwi42Z120KQQG8024Wbl2iUuUl1"))
	return tokenstring, err
}
func DecodeAccessToken(accestoken string) (models.User, error) {
	//TODO: implement JWT
	token, err := jwt.Parse(accestoken, func(token *jwt.Token) (interface{}, error) {
		return []byte("ZG6wkVwi42Z120KQQG8024Wbl2iUuUl1"), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		log.Fatal(err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		//unix timestamp compare with current time
		if time.Now().Before(time.Unix(int64(claims["exp"].(float64)), 0)) {
			return models.User{
				UID:     claims["sub"].(string),
				Picture: claims["picture"].(string),
				Name:    claims["name"].(string),
			}, nil
		}
	} else {
		log.Fatal(err)
	}

	return models.User{}, nil
}

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
		"client_id":     {os.Getenv("GOOGLE_OAUTH_CLIENT_ID")},
		"client_secret": {os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")},
		"redirect_uri":  {os.Getenv("GOOGLE_OAUTH_REDIRECT_URL")},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequest("POST", "https://www.googleapis.com/oauth2/v4/token", strings.NewReader(requestData.Encode()))
	if err != nil {
		log.Fatal(err)
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

	var tokenInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return err
	}

	//Call user info API with the access token as Bearer token
	userReq, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return err
	}
	userReq.Header.Add("Authorization", fmt.Sprintf("Bearer %v", tokenInfo["access_token"]))

	userResp, err := client.Do(userReq)
	if err != nil {
		return err
	}
	defer userResp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return err
	}

	c := ctx.Request().Context()
	clientOpts := options.Client().ApplyURI(
		fmt.Sprintf("%v", os.Getenv("DB_CONN")))
	mongoClient, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	var User models.User
	usermongo := mongoClient.Database("task-planner").Collection("users").FindOne(c, bson.M{"uid": userInfo["sub"].(string)})
	if err != nil {
		log.Fatal(err)
	}
	usermongo.Decode(&User)
	if User.Name == "" {
		User.Name = userInfo["name"].(string)
		User.Picture = userInfo["picture"].(string)
		User.UID = userInfo["sub"].(string)
		_, err = mongoClient.Database("task-planner").Collection("users").InsertOne(c, User)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err = mongoClient.Database("task-planner").Collection("users").UpdateOne(c, bson.M{"uid": User.UID}, bson.M{"$set": User})
		if err != nil {
			log.Fatal(err)
		}
	}

	generatedToken, err := generateAccessToken(User.UID, User.Picture, User.Name)
	if err != nil {
		log.Fatal(err)
		return ctx.Redirect(http.StatusFound, "/google-auth")
	}
	cookie := &http.Cookie{
		Name:    "access_token",
		Value:   generatedToken,
		Path:    "/",
		Expires: time.Now().Add(365 * 24 * time.Hour),
		Secure:  true,
	}
	//ctx.SetCookie(cookie)
	ctx.Response().Header().Set("Set-Cookie", cookie.String())

	return ctx.Redirect(http.StatusFound, "/")
}

func Logout(ctx echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "access_token"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	cookie.Path = "/"
	ctx.SetCookie(cookie)
	return ctx.Redirect(http.StatusFound, "/")
}

// Middleware to check if the user is logged in
func CheckLoggedIn(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if ctx.Request().URL.Path == "/oauth/callback" {
			return next(ctx)
		}
		if ctx.Request().URL.Path == "/google-auth" {
			return next(ctx)
		}
		// path contains api
		containsAPI := strings.Contains(ctx.Request().URL.Path, "api")
		if containsAPI {
			return next(ctx)
		}
		_, err := ctx.Cookie("access_token")
		// TODO: Check if the user is logged in is valid (check if the token is still valid)
		if err != nil {
			return ctx.Redirect(http.StatusFound, "/google-auth")
		}
		return next(ctx)
	}
}
