package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/girirock/task-planner/cmd/handlers"
	"github.com/girirock/task-planner/cmd/models"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type userData struct {
	Name    string
	Picture string
	Tasks   []models.Task
}

func main() {
	e := echo.New()
	godotenv.Load()

	e.Static("/js", "assets/js")
	e.Static("/css", "assets/css")
	e.Renderer = newTemplate()
	//use the process middleware to check if the user is logged in
	e.Use(handlers.CheckLoggedIn)
	e.GET("/", func(c echo.Context) error {
		accessToken, _ := c.Cookie("access_token")
		User, err := handlers.DecodeAccessToken(accessToken.Value)
		if err != nil {
			return c.Redirect(http.StatusFound, "/google-auth")
		}
		userData := userData{
			Name:    User.Name,
			Picture: User.Picture,
			Tasks:   []models.Task{},
		}
		return c.Render(200, "index", userData)
	})
	e.GET("/google-auth", handlers.CallGoogleOAuth)
	e.GET("/tasks", handlers.GetTasks)
	e.DELETE("/tasks", handlers.DeleteTask)
	e.GET("/oauth/callback", handlers.GoogleOAuthCallback)
	e.GET("/logout", handlers.Logout)
	e.Logger.Fatal(e.Start(":42069"))
}
