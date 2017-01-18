package main

import (
	"fmt"
	"go-tfts"
	"log"
	"net/http"
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

// GetUser action maps to route /users/{id:i}
func (uc UserController) GetUser(c rest.Context) rest.ResponseSender {
	return rest.NewOKJSONResponse(User{"John", "Smith"})
}

// GetAllUsers action maps to route /users (GET)
func (uc UserController) GetAllUsers(c rest.Context) rest.ResponseSender {
	return rest.NewGzipResponse(rest.NewOKJSONResponse(users))
}

// CreateUser action maps to route /users (POST)
func (uc UserController) CreateUser(c rest.Context) rest.ResponseSender {
	var user User

	c.BindJSONEntity(&user)

	users = append(users, user)

	uc.db.execute(fmt.Sprintf("INSERT INTO User (`firstName`, `lastName`) VALUES ('%v', '%v')", user.FirstName, user.LastName))

	return rest.NewCreatedJSONResponse(user)
}

func main() {
	root := "/api/v1"

	userController := UserController{stubDB{}}

	router := rest.DefaultRouter().Prefix(root).RouteMap(
		rest.NewRoute().For("/users/{id:i}").
			With(rest.MethodGET, userController.GetUser),
		rest.NewRoute().For("/users").
			With(rest.MethodGET, userController.GetAllUsers).
			And(rest.MethodPOST, userController.CreateUser),
	)

	h := rest.NewHandler(router)

	h.Use(func(c rest.Context) {
		c.Next()
	})

	h.Use(func(c rest.Context) {
		c.Next()
	})

	http.Handle(root+"/", h)

	log.Fatal(http.ListenAndServe(":10100", nil))
}
