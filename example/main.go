package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Mehokm/doze"
)

// User struct holds basic data about a user
type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

var users = []User{
	User{"John", "Smith"},
	User{"Jane", "Doe"},
	User{"Bruce", "Wayne"},
}

// This is a stub db struct
type stubDB struct{}

func (s stubDB) execute(query string) {
	fmt.Println("I executed: " + query)
}

// UserController is a basic struct to encapsulate all user actions
type UserController struct {
	db stubDB
}

// // RedirectToUser action maps to route /users/{id:i}/{to:i}
// func (uc UserController) RedirectToUser(c doze.Context) doze.ResponseSender {
// 	http.Redirect(c.ResponseWriter, c.Request.Request, "http://www.google.com", 301)
// 	return doze.NewNoContentResponse()
// }

// GetUser action maps to route /users/{id:i}
func (uc UserController) GetUser(c doze.Context) doze.ResponseSender {
	return doze.NewOKJSONResponse(User{"John", "Smith"})
}

// GetAllUsers action maps to route /users (GET)
func (uc UserController) GetAllUsers(c doze.Context) doze.ResponseSender {
	return doze.NewOKJSONResponse(users)
}

// CreateUser action maps to route /users (POST)
func (uc UserController) CreateUser(c doze.Context) doze.ResponseSender {
	var user User

	c.BindJSONEntity(&user)

	users = append(users, user)

	uc.db.execute(fmt.Sprintf("INSERT INTO User (`firstName`, `lastName`) VALUES ('%v', '%v')", user.FirstName, user.LastName))

	return doze.NewCreatedJSONResponse(user)
}

func main() {
	root := "/api/v1"

	userController := UserController{stubDB{}}

	router := doze.DefaultRouter().Prefix(root).RouteMap(
		doze.NewRoute().For("/users/{id:i}").
			With(http.MethodGet, userController.GetUser),
		// doze.NewRoute().For("/users/{id:i}/{to:i}").
		// 	With(http.MethodGet, userController.RedirectToUser),
		doze.NewRoute().For("/users").
			With(http.MethodGet, userController.GetAllUsers).
			And(http.MethodPost, userController.CreateUser),
	)

	h := doze.NewHandler(router)

	http.Handle(h.Pattern(), h)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
