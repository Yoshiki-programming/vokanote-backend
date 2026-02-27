package router

import (
	"github.com/Yoshiki-programming/vokanote-backend/inter/handler"
	"net/http"
)

func Router() {
	http.HandleFunc("/delete", handler.DeleteHandler)
	http.HandleFunc("/generateContent", handler.GenerateContentHandler)
}
