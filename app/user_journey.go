package app

import (
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
