package app

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	// "github.com/goforj/godump"
	"github.com/jmoiron/sqlx"
	"github.com/starfederation/datastar-go/datastar"
)

func indexPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		IndexPageTemplate().Render(r.Context(), w)
	}
}

func chooseCustomerHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ok, customers, _ := getHomepageData(db, w, r)
		if !ok {
			return
		}
		vm := PickCustomerVm{Customers: customers}
		ChooseCustomerTemplate(vm).Render(r.Context(), w)
	}
}

func chooseCustomerSubmitHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		customerId, ok := formValueAsIntOrErr(w, r, "customer_id")
		if !ok {
			return
		}

		if customerId == "0" {
			ok, customers, _ := getHomepageData(db, w, r)
			if !ok {
				return
			}
			vm := PickCustomerVm{Customers: customers, HasError: true}
			ChooseCustomerTemplate(vm).Render(r.Context(), w)
			return
		}
		updateJourneyCookie(w, r, map[string]string{
			"customer_id": customerId,
		})
		http.Redirect(w, r, "/visits/choose-location", http.StatusSeeOther)
	}
}

func chooseLocationHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		journey := readJourneyCookie(r)

		customerId := journey["customer_id"]
		if customerId == "" {
			errorHandler(w, r, "http: error reading customer_id from cookie path")
			return
		}

		ok, _, locations := getHomepageData(db, w, r)
		if !ok {
			return
		}

		if customerId == "0" {
			errorHandler(w, r, "http: tempered request")
			return
		}

		var customer Customer

		if err := db.GetContext(r.Context(), &customer, SelectCustomerByIdSql, customerId); err != nil {
			errorHandler(w, r, fmt.Sprintf("sql: error getting customer by id - %v", err))
			return
		}

		vm := PickLocationVm{
			CustomerName: customer.Name,
			Locations:    filteredLocations(locations, customerId),
			CustomerId:   customerId,
		}

		ChooseLocationTemplate(vm).Render(r.Context(), w)
	}
}

func chooseLocationSubmitHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		customerId, ok := formValueAsIntOrErr(w, r, "customer_id")
		if !ok {
			return
		}

		if customerId == "0" {
			errorHandler(w, r, "http: customer id not being set is 0")
			return
		}

		locationId, ok := formValueAsIntOrErr(w, r, "location_id")
		if !ok {
			return
		}

		if locationId == "0" {
			ok, _, locations := getHomepageData(db, w, r)
			if !ok {
				return
			}

			vm := PickLocationVm{Locations: filteredLocations(locations, customerId), HasError: true}
			vm.CustomerId = customerId
			ChooseLocationTemplate(vm).Render(r.Context(), w)
			return
		}

		updateJourneyCookie(w, r, map[string]string{
			"location_id": locationId,
		})

		http.Redirect(w, r, "/visits/log-visit", http.StatusSeeOther)
	}

}

func logVisitHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		journey := readJourneyCookie(r)

		fmt.Println(journey)

		locationId := journey["location_id"]
		if locationId == "" {
			errorHandler(w, r, "http: error reading location_id from cookie path")
			return
		}

		if locationId == "0" {
			errorHandler(w, r, "http: location_id is 0 this shouldn't happen")
			return
		}

		var loc struct {
			LocationName string `db:"location_name"`
			CustomerName string `db:"customer_name"`
			CustomerId   string `db:"customer_id"`
			LocationId   string `db:"location_id"`
		}

		if err := db.Get(&loc, SelectLocationById, locationId); err != nil {
			errorHandler(w, r, fmt.Sprintf("sql: error getting location - %v", err))
			return
		}

		vm := VisitVM{
			Date:          time.Now().Format("2006-01-02"),
			Duration:      "60",
			CustomerName:  loc.CustomerName,
			LocationName:  loc.LocationName,
			CustomerId:    loc.CustomerId,
			LocationId:    loc.LocationId,
			IsSubmission:  false,
			VisitVMErrors: VisitVMErrors{HasTimeError: false, HasDateError: false},
		}

		LogInfo("starting log a location")
		LogVisitTemplate(vm).Render(r.Context(), w)
	}
}

func logVisitSubmitHandler(db *sqlx.DB, uploadsDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("ok")
		r, err := parseMultipart(r)
		if err != nil {
			errorHandler(w, r, fmt.Sprintf("http: issue parsing multipart form - %v", err), 500)
			return
		}

		vm := validateVisitSubmission(r)

		if vm.HasErrors() {
			LogVisitTemplate(vm).Render(r.Context(), w)
			return
		}

		visitId, err := logVisitData(db, r, uploadsDir)
		if err != nil {
			errorHandler(w, r, err.Error())
			return
		}
		fmt.Println("visit id is: ", visitId)

		var imagePaths = []string{}

		if err := db.SelectContext(r.Context(), &imagePaths, `SELECT filename from images where visit_id = ?;`, visitId); err != nil {
			errorHandler(w, r, fmt.Sprintf("sql: error getting images - %v", err))
			return
		}

		fmt.Printf("there are %v images", len(imagePaths))

		// we also need times and dates and these are part of a forthcoming migration
		var visit struct {
			CustomerName string `db:"customer_name"`
			LocationName string `db:"location_name"`
			EmployeeName string `db:"employee_name"`
		}

		if err = db.GetContext(r.Context(), &visit, SelectVisitDataSql, visitId); err != nil {
			errorHandler(w, r, fmt.Sprintf("sql: error geting visit data: %v", err))
			return
		}

		cvm := ConfirmationVm{
			VisitId:      visitId,
			LocationName: visit.LocationName,
			EmployeeName: visit.EmployeeName,
			CustomerName: visit.CustomerName,
			ImagePaths:   imagePaths,
		}

		ConfirmationTemplate(cvm).Render(r.Context(), w)

	}
}

func validateVisitDateHandler(w http.ResponseWriter, r *http.Request) {
	ds := dateSignals{}
	datastar.ReadSignals(r, &ds)
	dateError := hasDateError(ds.VisitDate)
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(VisitDateInputTemplate(true, dateError))
}

func validateVisitTimeHandler(w http.ResponseWriter, r *http.Request) {
	ts := timeSignals{}
	datastar.ReadSignals(r, &ts)
	fmt.Println(ts)
	timeError := hasTimeError(ts.VisitTime)
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(VisitTimeInputTemplate(true, timeError))
}

func validateVisitNotesHandler(w http.ResponseWriter, r *http.Request) {
	ns := notesSignals{}
	datastar.ReadSignals(r, &ns)
	notesError := hasNotesError(ns.VisitNotes)
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(VisitNotesInputTemplate(true, notesError))
}

func errorHandler(w http.ResponseWriter, r *http.Request, msg string, statusCode ...int) {
	status := 500
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	_, file, line, ok := runtime.Caller(1) // 1 = caller of this function
	if ok {
		msg = fmt.Sprintf("%s (at %s:%d)", msg, file, line)
	}
	LogError(msg)
	w.WriteHeader(status)
	ErrorPageTemplate().Render(r.Context(), w)
}

func neuterDirectoryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func basicAuthHandler(user string, user_password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if ok {
				usernameHash := sha256.Sum256([]byte(username))
				passwordHash := sha256.Sum256([]byte(password))
				expectedUsernameHash := sha256.Sum256([]byte(user))
				expectedPasswordHash := sha256.Sum256([]byte(user_password))

				usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
				passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

				if usernameMatch && passwordMatch {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}
