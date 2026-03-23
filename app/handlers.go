package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
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

func chooseCustomerPost(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		customerId, ok := formValueAsIntOrErr(w, r, "customer_id")
		if !ok {
			return
		}

		if customerId == 0 {
			ok, customers, _ := getHomepageData(db, w, r)
			if !ok {
				return
			}
			vm := PickCustomerVm{Customers: customers, HasError: true}
			ChooseCustomer(vm).Render(r.Context(), w)
			return
		}

		url := fmt.Sprintf("/visits/%v/choose-location", customerId)
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

// ----------------------------------------------------------------------------------

func chooseLocationGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		customerId, ok := pathValueAsIntOrErr(w, r, "customer_id")
		if !ok {
			return
		}

		ok, _, locations := getHomepageData(db, w, r)
		if !ok {
			return
		}

		if customerId == 0 {
			renderServerError(w, r, "http: tempered request")
			return
		}

		vm := PickLocationVm{Locations: filteredLocations(locations, customerId), CustomerId: customerId}
		ChooseLocation(vm).Render(r.Context(), w)
	}
}

func chooseLocationPost(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// hidden value
		customerId, ok := formValueAsIntOrErr(w, r, "customer_id")
		if !ok {
			return
		}

		if customerId == 0 {
			renderServerError(w, r, "http: customer id not being set is 0")
			return
		}

		locationId, ok := formValueAsIntOrErr(w, r, "location_id")
		if !ok {
			return
		}

		if locationId == 0 {
			ok, _, locations := getHomepageData(db, w, r)
			if !ok {
				return
			}

			vm := PickLocationVm{Locations: filteredLocations(locations, customerId), HasError: true}
			vm.CustomerId = customerId
			ChooseLocation(vm).Render(r.Context(), w)
			return
		}

		url := fmt.Sprintf("/visits/%v/log-visit", locationId)
		http.Redirect(w, r, url, http.StatusSeeOther)
	}

}

func logVisitGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		locationId, ok := pathValueAsIntOrErr(w, r, "location_id")
		if !ok {
			return
		}

		if locationId == 0 {
			renderServerError(w, r, "http: location_id is 0 this shouldn't happen")
			return
		}

		loc := locationByCustomer{}

		if err := db.Get(&loc, SelectLocationById, locationId); err != nil {
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

		LogInfo("starting log a location")
		LogVisit(vm).Render(r.Context(), w)
	}
}

func logVisitPost(db *sqlx.DB, uploadsDir string) http.HandlerFunc {
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

		visitId, err := logVisit(db, r, uploadsDir)
		if err != nil {
			renderServerError(w, r, err.Error())
			return
		}

		time.Sleep(3 * time.Second)
		LogInfo("stage 2 finished and redirecting")

		url := fmt.Sprintf("/visits/%v/complete", visitId)
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

func confirmVisit(_ *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Thanks().Render(r.Context(), w)
	}
}
