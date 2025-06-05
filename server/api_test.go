package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterLoginListPresence(t *testing.T) {
	// Setup routes for test
	setupRoutes()

	// --- Register ---
	user := User{Username: "alice", Password: "secret"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	registerHandler(w, req)
	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("register failed: %v", w.Result().Status)
	}

	// --- Login ---
	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	loginHandler(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("login failed: %v", w.Result().Status)
	}

	// --- List Users ---
	req = httptest.NewRequest("GET", "/api/users", nil)
	w = httptest.NewRecorder()
	listUsersHandler(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("list users failed: %v", w.Result().Status)
	}
	var users []User
	if err := json.NewDecoder(w.Body).Decode(&users); err != nil || len(users) == 0 {
		t.Fatalf("list users decode failed: %v", err)
	}

	// --- Presence ---
	req = httptest.NewRequest("POST", "/api/presence?username=alice&online=true", nil)
	w = httptest.NewRecorder()
	presenceHandler(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("presence update failed: %v", w.Result().Status)
	}
}
