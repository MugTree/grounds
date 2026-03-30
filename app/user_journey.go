package app

import (
	"net/http"
	"net/url"
	"time"

	"github.com/goforj/godump"
)

const JourneyCookieName string = "visit_journey"

func deleteJourneyCookie(w http.ResponseWriter) {
	LogInfo("cookie: deleting the journey cookie")
	c := &http.Cookie{
		Name:     JourneyCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}

func journeyComplete(r *http.Request, secretKey []byte) (bool, error) {
	journey, err := readJourneyCookie(r, secretKey)

	godump.Dump(journey)

	if err != nil {
		return false, err
	}

	complete := journey["journey_complete"]

	if complete == "true" {
		LogInfo("Journey is complete!!!")
		return true, nil
	}
	LogInfo("Journey NOT complete!!!")
	return false, nil
}

func updateJourneyCookie(w http.ResponseWriter, r *http.Request, secretKey []byte, updates map[string]string) error {

	values := url.Values{}
	if _, err := r.Cookie(JourneyCookieName); err == nil {

		cookieVal, err := ReadSigned(r, JourneyCookieName, secretKey)
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

	return WriteSigned(w, cookie, secretKey)

}

func readJourneyCookie(r *http.Request, secretKey []byte) (map[string]string, error) {
	values := make(map[string]string)

	cookieVal, err := ReadSigned(r, JourneyCookieName, secretKey)
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
