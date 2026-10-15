package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLocalLoginMeAndLogout(t *testing.T) {
	_, _, api := newTestAPI(t)
	cookie := login(t, api)
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Secure ||
		cookie.Path != "/" || cookie.MaxAge != int(sessionTTL.Seconds()) || cookie.Expires.IsZero() {
		t.Fatalf("incorrect HTTP demo cookie flags: %+v", cookie)
	}
	response := request(api, http.MethodGet, "/api/v1/auth/me", "", cookie)
	if response.Code != http.StatusOK {
		t.Fatalf("me: %d %s", response.Code, response.Body.String())
	}
	var who identity
	if err := json.Unmarshal(response.Body.Bytes(), &who); err != nil || who.Login != demoLogin || who.AvatarURL != "" {
		t.Fatalf("unexpected identity: %s, %v", response.Body.String(), err)
	}
	logout := request(api, http.MethodPost, "/api/v1/auth/logout", "", cookie)
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout: %d", logout.Code)
	}
	deleted := logout.Result().Cookies()
	if len(deleted) != 1 || deleted[0].Name != sessionCookieName || deleted[0].MaxAge != -1 ||
		deleted[0].Value != "" || !deleted[0].HttpOnly || deleted[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("logout did not clear session cookie: %v", deleted)
	}
	if response := request(api, http.MethodGet, "/api/v1/auth/me", "", deleted[0]); response.Code != http.StatusUnauthorized {
		t.Fatalf("cleared cookie authenticated: %d", response.Code)
	}
	for _, path := range []string{"/api/v1/auth/login", "/api/v1/auth/logout"} {
		if response := request(api, http.MethodGet, path, "", cookie); response.Code != http.StatusMethodNotAllowed {
			t.Fatalf("legacy GET auth operation still active at %s: %d", path, response.Code)
		}
	}
}

func TestAuthenticationRequiredEveryDungeonRoute(t *testing.T) {
	t.Setenv("KROMBAT_TEST_USER", demoLogin)
	_, client, api := newTestAPI(t, testDungeon(t, "easy"))
	for _, route := range []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/auth/me"},
		{"GET", "/api/v1/dungeons"},
		{"POST", "/api/v1/dungeons"},
		{"GET", testPath},
		{"DELETE", testPath},
		{"POST", testPath + "/attacks"},
		{"POST", testPath + "/actions"},
		{"GET", testPath + "/resources"},
		{"POST", testPath + "/cel-eval"},
	} {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			req.Header.Set("X-Test-User", demoLogin)
			response := httptest.NewRecorder()
			api.ServeHTTP(response, req)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("missing cookie was not rejected: %d %s", response.Code, response.Body.String())
			}
		})
	}
	if len(client.Actions()) != 0 {
		t.Fatal("unauthenticated requests reached Kubernetes")
	}
}

func TestSignedSessionsRejectTamperingExpiryAndOtherUsers(t *testing.T) {
	auth, err := newSessionAuth(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(2000000000, 0)
	auth.now = func() time.Time { return now }
	payload := sessionPayload{Login: demoLogin, ExpiresAt: now.Add(time.Hour).Unix(), Nonce: "nonce"}
	token, err := auth.sign(payload)
	if err != nil || !auth.valid(token) {
		t.Fatalf("valid token rejected: %v", err)
	}
	otherSecret, _ := newSessionAuth(strings.Repeat("x", 32))
	otherSecret.now = auth.now
	if otherSecret.valid(token) {
		t.Fatal("token accepted with different secret")
	}
	for _, invalid := range []string{
		"", "not-a-token", token + ".", "x" + token, token[:len(token)-4],
		strings.Repeat("a", 4097), "bm90LWpzb24.invalid",
	} {
		if auth.valid(invalid) {
			t.Errorf("invalid token accepted: %q", invalid)
		}
	}
	for _, invalidPayload := range []sessionPayload{
		{Login: demoLogin, ExpiresAt: now.Unix(), Nonce: "expired"},
		{Login: demoLogin, ExpiresAt: now.Add(-time.Hour).Unix(), Nonce: "expired"},
		{Login: "other-user", ExpiresAt: payload.ExpiresAt, Nonce: "other"},
		{Login: demoLogin, ExpiresAt: payload.ExpiresAt},
	} {
		invalid, err := auth.sign(invalidPayload)
		if err != nil {
			t.Fatal(err)
		}
		if auth.valid(invalid) {
			t.Errorf("invalid signed payload accepted: %+v", invalidPayload)
		}
	}
}

func TestExplicitSessionSecretValidation(t *testing.T) {
	for _, secret := range []string{"", "short", strings.Repeat(" ", 32), strings.Repeat("s", 31)} {
		if _, err := newSessionAuth(secret); err == nil {
			t.Errorf("invalid SESSION_SECRET accepted: %q", secret)
		}
	}
	if _, err := newSessionAuth(strings.Repeat("s", 32)); err != nil {
		t.Fatal(err)
	}
	if _, err := New(nil, testSecret); err == nil {
		t.Fatal("missing Kubernetes client accepted")
	}
	t.Setenv("SESSION_SECRET", "")
	if _, err := newSessionAuth(testSecret); err != nil {
		t.Fatalf("constructor unexpectedly depended on environment or init: %v", err)
	}
}

func TestHealthLegacyRoutesAndRateLimits(t *testing.T) {
	_, _, api := newTestAPI(t)
	response := request(api, http.MethodGet, "/healthz", "", nil)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("health: %d %s", response.Code, response.Body.String())
	}
	if response.Header().Get("X-Content-Type-Options") != "nosniff" ||
		response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected response headers are missing")
	}
	for _, path := range []string{
		"/api/v1/auth/callback", "/api/v1/auth/test-login", "/api/v1/events",
		"/api/v1/leaderboard", "/api/v1/profile", "/anything-else",
	} {
		if response := request(api, http.MethodGet, path, "", nil); response.Code != http.StatusNotFound {
			t.Errorf("legacy route %s is still active: %d", path, response.Code)
		}
	}
	cookie := login(t, api)
	response = request(api, http.MethodPost, "/api/v1/auth/login", "", nil)
	if response.Code != http.StatusTooManyRequests || response.Header().Get("Retry-After") == "" {
		t.Fatalf("login rate limit: %d %s", response.Code, response.Body.String())
	}
	cookie.Value += "tampered"
	if response := request(api, http.MethodGet, "/api/v1/auth/me", "", cookie); response.Code != http.StatusUnauthorized {
		t.Fatal("tampered cookie authenticated")
	}
}
