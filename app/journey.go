package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"time"
)

const JourneyCookieName string = "visit_journey"

func deleteJourneyCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     JourneyCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	LogInfo("cookie: deleting the journey cookie")
}

func updateJourneyCookie(w http.ResponseWriter, r *http.Request, secretKey []byte, updates map[string]string) error {

	values := url.Values{}
	if _, err := r.Cookie(JourneyCookieName); err == nil {

		cookieVal, err := readSignedCookie(r, JourneyCookieName, secretKey)
		if err != nil {
			return err
		}

		existing, _ := url.ParseQuery(cookieVal)
		for k, v := range existing {
			if len(v) > 0 {
				values[k] = v
			}
		}
	}

	for k, v := range updates {
		if v != "" {
			values.Set(k, v)
		}
	}

	cookie := http.Cookie{
		Name:     JourneyCookieName,
		Value:    values.Encode(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	}

	return writeSignedCookie(w, cookie, secretKey)

}

func readJourneyCookie(r *http.Request, secretKey []byte) (map[string]string, error) {
	values := make(map[string]string)

	cookieVal, err := readSignedCookie(r, JourneyCookieName, secretKey)
	if err != nil {
		return nil, err
	}

	parsed, _ := url.ParseQuery(cookieVal)

	for k, v := range parsed {
		if len(v) > 0 {
			values[k] = v[0]
		}
	}

	return values, nil
}

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
)

func writeCookie(w http.ResponseWriter, cookie http.Cookie) error {
	// Encode the cookie value using base64.
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))

	// Check the total length of the cookie contents. Return the ErrValueTooLong
	// error if it's more than 4096 bytes.
	if len(cookie.String()) > 4096 {
		return ErrValueTooLong
	}

	// Write the cookie as normal.
	http.SetCookie(w, &cookie)

	return nil
}

func readCookie(r *http.Request, name string) (string, error) {
	// Read the cookie as normal.
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	// Decode the base64-encoded cookie value. If the cookie didn't contain a
	// valid base64-encoded value, this operation will fail and we return an
	// ErrInvalidValue error.
	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", ErrInvalidValue
	}

	// Return the decoded cookie value.
	return string(value), nil
}

func writeSignedCookie(w http.ResponseWriter, cookie http.Cookie, secretKey []byte) error {
	// Calculate a HMAC signature of the cookie name and value, using SHA256 and
	// a secret key (which we will create in a moment).
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)

	// Prepend the cookie value with the HMAC signature.
	cookie.Value = string(signature) + cookie.Value

	// Call our Write() helper to base64-encode the new cookie value and write
	// the cookie.
	return writeCookie(w, cookie)
}

func readSignedCookie(r *http.Request, name string, secretKey []byte) (string, error) {
	// Read in the signed value from the cookie. This should be in the format
	// "{signature}{original value}".
	signedValue, err := readCookie(r, name)
	if err != nil {
		return "", err
	}

	// A SHA256 HMAC signature has a fixed length of 32 bytes. To avoid a potential
	// 'index out of range' panic in the next step, we need to check sure that the
	// length of the signed cookie value is at least this long. We'll use the
	// sha256.Size constant here, rather than 32, just because it makes our code
	// a bit more understandable at a glance.
	if len(signedValue) < sha256.Size {
		return "", ErrInvalidValue
	}

	// Split apart the signature and original cookie value.
	signature := signedValue[:sha256.Size]
	value := signedValue[sha256.Size:]

	// Recalculate the HMAC signature of the cookie name and original value.
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	// Check that the recalculated signature matches the signature we received
	// in the cookie. If they match, we can be confident that the cookie name
	// and value haven't been edited by the client.
	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", ErrInvalidValue
	}

	// Return the original cookie value.
	return value, nil
}
