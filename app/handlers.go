package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfederation/datastar-go/datastar"
)

func indexPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		IndexPage().Render(r.Context(), w)
	}
}

type JourneyData struct {
	CustomerId int
	LocationId int
}

const JourneyDataCookieName string = "journeydata"

func chooseCustomer(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ok, customers, _ := getHomepageData(db, w, r)
		if !ok {
			return
		}
		vm := PickCustomerVm{Customers: customers}
		ChooseCustomer(vm).Render(r.Context(), w)
	}
}

func chooseCustomerSubmit(db *sqlx.DB) http.HandlerFunc {
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

func chooseLocation(db *sqlx.DB) http.HandlerFunc {
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

		var customer Customer

		if err := db.GetContext(r.Context(), &customer, SelectCustomerByIdSql, customerId); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting customer by id - %v", err))
			return
		}

		vm := PickLocationVm{
			CustomerName: customer.Name,
			Locations:    filteredLocations(locations, customerId),
			CustomerId:   customerId,
		}

		ChooseLocation(vm).Render(r.Context(), w)
	}
}

func choosteLocationSubmit(db *sqlx.DB) http.HandlerFunc {
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

func logVisit(db *sqlx.DB) http.HandlerFunc {
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
			CustomerId:   loc.CustomerId,
			LocationId:   loc.LocationId,
		}

		LogInfo("starting log a location")
		LogVisit(vm).Render(r.Context(), w)
	}
}

// will use datastar to make the form submission then i can return some HMTL into the page
func logVisitSubmit(db *sqlx.DB, uploadsDir string) http.HandlerFunc {
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

		visitId, err := logVisitData(db, r, uploadsDir)

		fmt.Println("visit id is: ", visitId)

		if err != nil {
			renderServerError(w, r, err.Error())
			return
		}

		LogInfo("logVisitSubmit")

		sse := datastar.NewSSE(w, r)

		// data will be in an odd shape but complete

		simplerSqlDisentangle := `
		SELECT
      		*
		FROM visits v
		INNER JOIN employee e ON e.id = v.employee_id
		WHERE v.id = ?;
		SELECT* from images where visit_id = ?`

		fmt.Println(simplerSqlDisentangle)

		vm := ConfirmationVm{}
		//vm.Date

		sse.PatchElementTempl(Confirmation(vm))

	}
}

func logVisitConfirm(_ *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		visitId := r.FormValue("visit_id")

		fmt.Println(visitId)

		// update the visit

	}
}
