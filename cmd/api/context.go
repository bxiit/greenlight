package main

import (
	"context"
	"github.com/bxiit/greenlight/internal/data"
	"net/http"
)

type contextKey string

// Convert the string "user" to a contextKey type and assign it to the userContextKey
// constant. We'll use this constant as the key for getting and setting user information
// in the request context.
const userContextKey = contextKey("user")
const userInfoContextKey = contextKey("userInfo")

// The contextSetUser() method returns a new copy of the request with the provided
// User struct added to the context. Note that we use our userContextKey constant as the
// key.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextSetUserInfo(r *http.Request, userInfo *data.UserInfo) *http.Request {
	ctx := context.WithValue(r.Context(), userInfoContextKey, userInfo)
	return r.WithContext(ctx)
}

// The contextGetUser() retrieves the User struct from the request context. The only
// time that we'll use this helper is when we logically expect there to be User struct
// value in the context, and if it doesn't exist it will firmly be an 'unexpected' error.
// As we discussed earlier in the book, it's OK to panic in those circumstances.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}

func (app *application) contextGetUserInfo(r *http.Request) *data.UserInfo {
	userInfo, ok := r.Context().Value(userInfoContextKey).(*data.UserInfo)
	if !ok {
		panic("missing userInfo value in request context")
	}
	return userInfo
}
