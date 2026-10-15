package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	demoLogin         = "demo"
	sessionCookieName = "krombat_session"
	sessionTTL        = 4 * time.Hour
)

type identity struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatarUrl"`
}

type sessionPayload struct {
	Login     string `json:"login"`
	ExpiresAt int64  `json:"expiresAt"`
	Nonce     string `json:"nonce"`
}

type sessionAuth struct {
	secret []byte
	now    func() time.Time
}

func ValidateSessionSecret(secret string) error {
	if len(strings.TrimSpace(secret)) < 32 {
		return errors.New("SESSION_SECRET must contain at least 32 non-padding bytes; generate it during local setup")
	}
	return nil
}

func newSessionAuth(secret string) (*sessionAuth, error) {
	if err := ValidateSessionSecret(secret); err != nil {
		return nil, err
	}
	return &sessionAuth{secret: []byte(secret), now: time.Now}, nil
}

func (a *sessionAuth) sign(payload sessionPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(data)
	mac := hmac.New(sha256.New, a.secret)
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (a *sessionAuth) valid(token string) bool {
	if len(token) > 4096 {
		return false
	}
	encoded, signature, ok := strings.Cut(token, ".")
	if !ok {
		return false
	}
	sig, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, a.secret)
	_, _ = mac.Write([]byte(encoded))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return false
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}
	var payload sessionPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return false
	}
	return payload.Login == demoLogin && payload.Nonce != "" && payload.ExpiresAt > a.now().Unix()
}

func (a *sessionAuth) require(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !a.valid(cookie.Value) {
			writeError(w, "authentication required; use the local demo login", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	expires := h.auth.now().Add(sessionTTL)
	token, err := h.auth.sign(sessionPayload{
		Login: demoLogin, ExpiresAt: expires.Unix(), Nonce: rand.Text(),
	})
	if err != nil {
		writeError(w, "could not create session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Value: token, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Secure: false,
		MaxAge: int(sessionTTL.Seconds()), Expires: expires,
	})
	writeJSON(w, http.StatusOK, identity{Login: demoLogin, AvatarURL: ""})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, identity{Login: demoLogin, AvatarURL: ""})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Secure: false,
		MaxAge: -1, Expires: time.Unix(1, 0),
	})
	w.WriteHeader(http.StatusNoContent)
}
