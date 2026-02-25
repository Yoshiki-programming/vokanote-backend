package router

import (
	"github.com/Yoshiki-programming/vokanote-backend.git/inter/handler"
	"net/http"
)

func Router() {
	http.HandleFunc("/deleteUser", handler.DeleteUserHandler)
}
