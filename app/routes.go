package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/mugtree/grounds/app/db"
	"github.com/starfederation/datastar-go/datastar"
)

type dateSignals struct {
	VisitDate string `json:"visit_date"`
}

type timeSignals struct {
	VisitTime string `json:"visit_time"`
}

type notesSignals struct {
	VisitNotes string `json:"visit_notes"`
}

func setupRoutes(r chi.Router, queries *db.Queries, sqldb *sql.DB, session *scs.SessionManager, uploadsDir string) {

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

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
	})
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

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

	})

	r.Route("/visit", func(r chi.Router) {
		r.Get("/{visit_id}", func(w http.ResponseWriter, r *http.Request) {

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
		})
		r.Get("/step-1/", func(w http.ResponseWriter, r *http.Request) {

			ok, customers, _ := getHomepageData(queries, w, r)
			if !ok {
				return
			}
			ChooseCustomerTemplate(customers).Render(r.Context(), w)
		})
		r.Post("/step-1/", func(w http.ResponseWriter, r *http.Request) {

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
		})
		r.Get("/step-2/", func(w http.ResponseWriter, r *http.Request) {

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
		})
		r.Post("/step-2/", func(w http.ResponseWriter, r *http.Request) {

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
		})
		r.Route("/step-3", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {

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
			})
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {

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

			})
			//r.Get("/confirm", visitConfirmationHandler(db, sessions))
			r.Post("/validate-date", func(w http.ResponseWriter, r *http.Request) {
				ds := dateSignals{}
				datastar.ReadSignals(r, &ds)
				dateError := hasDateError(ds.VisitDate)
				sse := datastar.NewSSE(w, r)
				sse.PatchElementTempl(VisitDateInputTemplate(true, dateError))
			})
			r.Post("/validate-notes", func(w http.ResponseWriter, r *http.Request) {
				ns := notesSignals{}
				datastar.ReadSignals(r, &ns)
				notesError := hasNotesError(ns.VisitNotes)
				sse := datastar.NewSSE(w, r)
				sse.PatchElementTempl(VisitNotesInputTemplate(true, notesError))
			})
			r.Post("/validate-time", func(w http.ResponseWriter, r *http.Request) {
				ts := timeSignals{}
				datastar.ReadSignals(r, &ts)
				timeError := hasTimeError(ts.VisitTime)
				sse := datastar.NewSSE(w, r)
				sse.PatchElementTempl(VisitTimeInputTemplate(true, timeError))
			})
		})

	})
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
