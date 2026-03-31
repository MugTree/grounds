package app

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	// "github.com/goforj/godump"

	"github.com/jmoiron/sqlx"
	"github.com/starfederation/datastar-go/datastar"
)

func indexPageHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		deleteJourneyCookie(w)

		IndexPageTemplate().Render(r.Context(), w)
	}
}

func stepOneHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ok, customers, _ := getHomepageData(db, w, r)
		if !ok {
			return
		}
		ChooseCustomerTemplate(customers).Render(r.Context(), w)
	}
}

func stepOneSubmitHandler(db *sqlx.DB, cookieKey []byte) http.HandlerFunc {
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
			ChooseCustomerTemplate(customers).Render(r.Context(), w)
			return
		}

		updateJourneyCookie(w, r, cookieKey, map[string]string{"customer_id": customerId})

		http.Redirect(w, r, "/visits/choose-location", http.StatusSeeOther)
	}
}

type PickLocationVm struct {
	CustomerId   string
	CustomerName string
	Locations    []Location
	HasError     bool
}

func stepTwoHandler(db *sqlx.DB, cookieKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		journey, err := readJourneyCookie(r, cookieKey)
		if err != nil {
			errorHandler(w, r, fmt.Sprintf("http: read journey error %v", err))
		}

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

		ChooseLocationTemplate(filteredLocations(locations, customerId), customerId, customer.Name).Render(r.Context(), w)
	}
}

func stepTwoSubmitHandler(_ *sqlx.DB, cookieKey []byte) http.HandlerFunc {
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

		updateJourneyCookie(w, r, cookieKey, map[string]string{
			"location_id": locationId,
		})

		http.Redirect(w, r, "/visits/log-visit/", http.StatusSeeOther)
	}
}

type VisitVM struct {
	Date         string
	Time         string
	Duration     string
	Notes        string
	CustomerId   string
	CustomerName string
	LocationName string
	LocationId   string
	IsComplete   bool
	IsSubmission bool
	VisitVMErrors
}

func (v VisitVM) HasErrors() bool {
	if v.HasDateError || v.HasTimeError {
		return true
	}
	return false
}

type VisitVMErrors struct {
	HasTimeError  bool
	HasDateError  bool
	HasNotesError bool
}

func stepThreeHandler(db *sqlx.DB, cookieKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		journey, err := readJourneyCookie(r, cookieKey)
		if err != nil {
			errorHandler(w, r, fmt.Sprintf("http: read journey error %v", err))
		}

		// stop users going back to the form after submitting
		// ---------------------------------------------
		if journey["journey_complete"] == "true" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

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

		LogVisitTemplate(vm).Render(r.Context(), w)
	}
}

func stepThreeSubmitHandler(db *sqlx.DB, uploadsDir string, cookieKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// stop users going back to the form after submitting
		// ---------------------------------------------
		journey, err := readJourneyCookie(r, cookieKey)
		if journey["journey_complete"] == "true" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		r, err = parseMultipart(r)
		if err != nil {
			errorHandler(w, r, fmt.Sprintf("http: issue parsing multipart form - %v", err), 500)
			return
		}

		vm := validateVisit(r)
		if vm.HasErrors() {
			LogVisitTemplate(vm).Render(r.Context(), w)
			return
		}

		visitId, err := logVisitData(db, r, uploadsDir)
		if err != nil {
			errorHandler(w, r, err.Error())
			return
		}

		updateJourneyCookie(w, r, cookieKey, map[string]string{
			"visit_id":         strconv.Itoa(int(visitId)),
			"journey_complete": "true",
		})
		http.Redirect(w, r, "/visits/log-visit/complete", http.StatusSeeOther)
	}
}

type VisitCompleteVm struct {
	LocationName string
	CustomerName string
	EmployeeName string
	VisitId      string
	Time         string
	Date         string
	Duration     string
	ImagePaths   []string
}

func confirmationHandler(db *sqlx.DB, cookieKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// dont want this page hit directly
		journey, err := readJourneyCookie(r, cookieKey)
		if journey["journey_complete"] == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		var imagePaths = []string{}
		if err != nil {
			errorHandler(w, r, fmt.Sprintf("http: read journey error %v", err))
		}

		visitId := journey["visit_id"]
		if err := db.SelectContext(r.Context(), &imagePaths,
			`SELECT filename from images where visit_id = ?;`, visitId); err != nil {
			errorHandler(w, r, fmt.Sprintf("sql: error getting images - %v", err))
			return
		}

		var visit struct {
			CustomerName string `db:"customer_name"`
			LocationName string `db:"location_name"`
			EmployeeName string `db:"employee_name"`
		}

		if err := db.GetContext(r.Context(), &visit, SelectVisitDataSql, visitId); err != nil {
			errorHandler(w, r, fmt.Sprintf("sql: error geting visit data: %v", err))
			return
		}

		cvm := VisitCompleteVm{
			VisitId:      visitId,
			LocationName: visit.LocationName,
			EmployeeName: visit.EmployeeName,
			CustomerName: visit.CustomerName,
			ImagePaths:   imagePaths,
		}

		VisitCompleteTemplate(cvm).Render(r.Context(), w)
	}
}

type dateSignals struct {
	VisitDate string `json:"visit_date"`
}

func validateVisitDateHandler(w http.ResponseWriter, r *http.Request) {
	ds := dateSignals{}
	datastar.ReadSignals(r, &ds)
	dateError := hasDateError(ds.VisitDate)
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(VisitDateInputTemplate(true, dateError))
}

type timeSignals struct {
	VisitTime string `json:"visit_time"`
}

func validateVisitTimeHandler(w http.ResponseWriter, r *http.Request) {
	ts := timeSignals{}
	datastar.ReadSignals(r, &ts)
	timeError := hasTimeError(ts.VisitTime)
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(VisitTimeInputTemplate(true, timeError))
}

type notesSignals struct {
	VisitNotes string `json:"visit_notes"`
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
