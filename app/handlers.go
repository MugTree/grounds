package app

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mugtree/grounds/app/db"

	// "github.com/goforj/godump"

	"github.com/alexedwards/scs/v2"
	"github.com/starfederation/datastar-go/datastar"
)

func visitPageHandler(queries *db.Queries, _ *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		vid, ok := pathValueAsIntOrErr(w, r, "visit_id")
		if !ok {
			return
		}

		d, err := queries.GetVisitById(ctx, vid)
		if err != nil {
			errorHandler(w, r, err.Error(), 500)
			return
		}

		t, err := time.Parse(time.RFC3339, d.VisitDatetime)
		if err != nil {
			errorHandler(w, r, err.Error(), 500)
			return
		}

		ViewVisitTemplate(
			d.CustomerName,
			d.LocationName,
			fmt.Sprintf("%v minutes", d.VisitDuration),
			t.Format("Mon Jan _2"),
		).Render(ctx, w)
	}

}

/*
q: what if user is already logged in and hits this ?
q: do we redirect non logins to this ?
q: do logged in get redirected to hp
*/

func loginHandler(_ *db.Queries, _ *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "GET":
			LoginTemplate().Render(r.Context(), w)
			return
		case "POST":

			// https://github.com/alexedwards/scs/tree/master/sqlite3store

			// use the sessions handler to create a user session

			name := r.FormValue("name")
			password := r.FormValue("password")

			fmt.Println(name, " ", password)

		default:
			errorHandler(w, r, fmt.Sprintf("%v - method not supported", r.Method), 404)
		}

	}
}

func indexPageHandler(queries *db.Queries, session *scs.SessionManager) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		vbe, err := queries.GetVisitsByEmployee(ctx, 1)
		if err != nil {
			errorHandler(w, r, err.Error())
			return
		}

		message := ""

		if session.Exists(ctx, "visit_complete") {
			// remove all the parts we dont need
			// --------------------------------------------------
			session.Remove(r.Context(), "customer_id")
			session.Remove(r.Context(), "location_id")
			session.Remove(r.Context(), "visit_complete")

			message = "We've added your visit..."

		}

		IndexPageTemplate(vbe, message).Render(r.Context(), w)
	}
}

func visitStepOneHandler(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ok, customers, _ := getHomepageData(queries, w, r)
		if !ok {
			return
		}
		ChooseCustomerTemplate(customers).Render(r.Context(), w)
	}
}

func visitStepOneSubmitHandler(queries *db.Queries, session *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		customerId, ok := formValueAsIntOrErr(w, r, "customer_id")
		if !ok {
			return
		}

		if customerId == 0 {
			ok, customers, _ := getHomepageData(queries, w, r)
			if !ok {
				return
			}
			ChooseCustomerTemplate(customers).Render(ctx, w)
			return
		}

		session.Put(r.Context(), "customer_id", customerId)
		http.Redirect(w, r, "/visit/step-2/", http.StatusSeeOther)
	}
}

func visitStepTwoHandler(queries *db.Queries, session *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		customerID := session.GetInt64(ctx, "customer_id")
		if customerID == 0 {
			errorHandler(w, r, "http: error reading customer_id from cookie path")
			return
		}

		ok, _, locations := getHomepageData(queries, w, r)
		if !ok {
			return
		}

		if customerID == 0 {
			errorHandler(w, r, "http: tampered request?")
			return
		}

		customer, err := queries.GetCustomerById(ctx, customerID)

		if err != nil {
			errorHandler(w, r, fmt.Sprintf("sql: error getting customer by id - %v", err))
			return
		}

		filteredLocations := func(locations []db.Location, customerId int64) []db.Location {
			filtered := make([]db.Location, 0, len(locations))
			for _, loc := range locations {
				if loc.CustomerID == customerId {
					filtered = append(filtered, loc)
				}
			}
			return filtered
		}

		ChooseLocationTemplate(filteredLocations(locations, customerID), customerID, customer.Name).Render(ctx, w)
	}
}

func visitStepTwoSubmitHandler(_ *db.Queries, session *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		customerId, ok := formValueAsIntOrErr(w, r, "customer_id")
		if !ok {
			return
		}

		if customerId == 0 {
			errorHandler(w, r, "http: customer id not being set is 0")
			return
		}

		locationId, ok := formValueAsIntOrErr(w, r, "location_id")
		if !ok {
			return
		}

		session.Put(r.Context(), "location_id", locationId)
		http.Redirect(w, r, "/visit/step-3/", http.StatusSeeOther)
	}
}

type VisitVM struct {
	Date         string
	Time         string
	Duration     string
	Notes        string
	CustomerId   int64
	CustomerName string
	LocationName string
	LocationId   int64
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

func visitStepThreeHandler(queries *db.Queries, session *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		locationId := session.GetInt64(ctx, "location_id")
		if locationId == 0 {
			errorHandler(w, r, "http: error reading location_id from cookie path")
			return
		}

		if locationId == 0 {
			errorHandler(w, r, "http: location_id is 0 this shouldn't happen")
			return
		}

		loc, err := queries.GetLocationById(ctx, locationId) //selectLocationData(r.Context(), db, locationId)
		if err != nil {
			errorHandler(w, r, err.Error())
			return
		}

		vm := VisitVM{
			Date:          time.Now().Format("2006-01-02"),
			Duration:      "60",
			CustomerName:  loc.CustomerName,
			LocationName:  loc.LocationName,
			CustomerId:    loc.CustomerID,
			LocationId:    loc.LocationID,
			IsSubmission:  false,
			VisitVMErrors: VisitVMErrors{HasTimeError: false, HasDateError: false},
		}

		LogVisitTemplate(vm).Render(ctx, w)
	}
}

func visitStepThreeSubmitHandler(queries *db.Queries, sqldb *sql.DB, uploadsDir string, session *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		r, err := parseMultipart(r)
		if err != nil {
			errorHandler(w, r, fmt.Sprintf("http: issue parsing multipart form - %v", err), 500)
			return
		}

		vm, err := validateVisit(r)
		if err != nil {
			errorHandler(w, r, err.Error())
			return
		}

		if vm.HasErrors() {
			LogVisitTemplate(vm).Render(ctx, w)
			return
		}

		visitId, err := logVisitData(queries, sqldb, r, uploadsDir)
		if err != nil {
			errorHandler(w, r, err.Error())
			return
		}

		session.Put(r.Context(), "visit_complete", strconv.Itoa(int(visitId)))
		http.Redirect(w, r, "/", http.StatusSeeOther)

	}
}

// type VisitCompleteVm struct {
// 	LocationName string
// 	CustomerName string
// 	EmployeeName string
// 	VisitId      string
// 	Time         string
// 	Date         string
// 	Duration     string
// 	ImagePaths   []string
// }

// func visitConfirmationHandler(db *sqlx.DB, session *scs.SessionManager) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		// dont want this page hit directly
// 		// ----------------------------------------------------
// 		if session.Get(r.Context(), "journey_complete") == "" {
// 			http.Redirect(w, r, "/", http.StatusSeeOther)
// 			return
// 		}

// 		var imagePaths = []string{}
// 		visitId := session.GetString(r.Context(), "visit_id")

// 		if err := db.SelectContext(r.Context(), &imagePaths,
// 			SelectImagePathsSql, visitId); err != nil {
// 			errorHandler(w, r, fmt.Sprintf("sql: error getting images - %v", err))
// 			return
// 		}

// 		visit := visitData{}
// 		if err := db.GetContext(r.Context(), &visit, SelectVisitDataSql, visitId); err != nil {
// 			errorHandler(w, r, fmt.Sprintf("sql: error geting visit data: %v", err))
// 			return
// 		}

// 		cvm := VisitCompleteVm{
// 			VisitId:      visitId,
// 			LocationName: visit.LocationName,
// 			EmployeeName: visit.EmployeeName,
// 			CustomerName: visit.CustomerName,
// 			ImagePaths:   imagePaths,
// 		}

// 		VisitCompleteTemplate(cvm).Render(r.Context(), w)
// 	}
// }

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

func getHomepageData(queries *db.Queries, w http.ResponseWriter, r *http.Request) (bool, []db.Customer, []db.Location) {

	ctx := r.Context()

	cust, err := queries.ListCustomers(ctx)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("sql: error getting customers: %v", err))
		return false, []db.Customer{}, []db.Location{}
	}

	loc, err := queries.ListLocations(ctx)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("sql: error getting locations: %v", err))
		return false, []db.Customer{}, []db.Location{}
	}

	return true, cust, loc

}

func pathValueAsIntOrErr(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {

	formVal := r.PathValue(key)

	if formVal == "" {
		errorHandler(w, r, fmt.Sprintf("http: incorrect path value %s on page %v", key, r.URL.Path))
		return 0, false
	}

	val, err := strconv.ParseInt(formVal, 10, 64)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("http: incorrect path value %v, should be numeric - on page %v", formVal, r.URL.Path))
		return 0, false
	}

	return val, true

}

func formValueAsIntOrErr(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {

	formVal := r.FormValue(key)

	if formVal == "" {
		errorHandler(w, r, fmt.Sprintf("http: incorrect form value %s on page %v", key, r.URL.Path))
		return 0, false
	}

	val, err := strconv.ParseInt(formVal, 10, 64)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("http: incorrect form value %v, should be numeric - on page %v", formVal, r.URL.Path))
		return 0, false
	}

	return val, true

}

func parseMultipart(r *http.Request) (*http.Request, error) {

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			return r, err
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			return r, err
		}
	}

	return r, nil
}
