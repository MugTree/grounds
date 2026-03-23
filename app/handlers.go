package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goforj/godump"
	"github.com/jmoiron/sqlx"
	"github.com/starfederation/datastar-go/datastar"
)

type JourneyData struct {
	CustomerId int
	LocationId int
}

const JourneyDataCookieName string = "journeydata"

func chooseCustomerGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ok, customers, _ := getHomepageData(db, w, r)
		if !ok {
			return
		}
		vm := PickCustomerVm{Customers: customers}
		ChooseCustomer(vm).Render(r.Context(), w)
	}
}

func chooseCustomerPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		customerId, ok := formValueAsIntOrErr(w, r, "customer_id")
		if !ok {
			return
		}

		journeyData := JourneyData{CustomerId: customerId}
		b, _ := json.Marshal(journeyData)

		http.SetCookie(w, &http.Cookie{
			Name:  JourneyDataCookieName,
			Value: url.QueryEscape(string(b)),
			Path:  "/",
		})

		http.Redirect(w, r, "/visits/choose-location", http.StatusSeeOther)
	}
}

// ----------------------------------------------------------------------------------

func chooseLocationGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, _, locations := getHomepageData(db, w, r)
		if !ok {
			return
		}

		cookie, err := r.Cookie(JourneyDataCookieName)
		if err != nil {
			renderServerError(w, r, fmt.Sprintf("http: ERROR HERE: %v", err))
			return
		}
		decoded, _ := url.QueryUnescape(cookie.Value)

		var data JourneyData
		if err := json.Unmarshal([]byte(decoded), &data); err != nil {
			renderServerError(w, r, fmt.Sprintf("http: error unmarshalling journey data %v", decoded))
		}

		godump.Dump(data)

		vm := PickLocationVm{Locations: filteredLocations(locations, data.CustomerId)}
		ChooseLocation(vm).Render(r.Context(), w)
	}
}

func chooseLocationPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("BEING HIT!!!!")

		locationId, ok := formValueAsIntOrErr(w, r, "location_id")
		if !ok {
			return
		}

		cookie, _ := r.Cookie(JourneyDataCookieName)
		decoded, _ := url.QueryUnescape(cookie.Value)

		var data JourneyData
		if err := json.Unmarshal([]byte(decoded), &data); err != nil {
			renderServerError(w, r, fmt.Sprintf("http: error unmarshalling journey data %v", decoded))
		}
		data.LocationId = locationId
		b, _ := json.Marshal(data)

		godump.Dump(data)

		http.SetCookie(w, &http.Cookie{
			Name:  "journey-data",
			Value: url.QueryEscape(string(b)),
			Path:  "/",
		})

		http.Redirect(w, r, "/visits/log-visit", http.StatusSeeOther)
	}

}

func logLocationGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		cookie, _ := r.Cookie("journey-data")
		decoded, _ := url.QueryUnescape(cookie.Value)

		var data JourneyData
		if err := json.Unmarshal([]byte(decoded), &data); err != nil {
			renderServerError(w, r, fmt.Sprintf("http: error unmarshalling journey data %v", decoded))
		}

		godump.Dump("journey data is", data)

		loc := locationByCustomer{}

		if err := db.Get(&loc, SelectLocationById, data.LocationId); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting location - %v", err))
			return
		}

		vm := visitVM{
			Date:         time.Now().Format("2006-01-02"),
			Duration:     "60",
			CustomerName: loc.CustomerName,
			LocationName: loc.LocationName,
			LocationId:   loc.LocationId,
		}

		//godump.Dump(vm)

		LogInfo("starting log a location")
		LogVisit(vm).Render(r.Context(), w)
	}
}

func logLocationPost(db *sqlx.DB, uploadsDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "multipart/form-data") {
			err := r.ParseMultipartForm(10 << 20)
			if err != nil {
				renderServerError(w, r, fmt.Sprintf("http: multipart form error - %v", err))
			}
		} else {
			err := r.ParseForm()
			if err != nil {
				renderServerError(w, r, fmt.Sprintf("http: form parse error - %v", err))
			}
		}

		sse := datastar.NewSSE(w, r)
		if err := logVisit(db, r, uploadsDir); err != nil {
			renderServerError(w, r, err.Error())
			sse.PatchElementTempl(VisitError())
			return
		}

		time.Sleep(3 * time.Second)
		sse.PatchElementTempl(Thanks())
		LogInfo("stage 2 finished and redirecting")
	}
}
