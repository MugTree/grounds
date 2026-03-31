package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"io"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/image/draw"
)

func filteredLocations(locations []Location, customerId string) []Location {
	filtered := make([]Location, 0, len(locations))
	for _, loc := range locations {
		if loc.CustomerId == customerId {
			filtered = append(filtered, loc)
		}
	}
	return filtered
}

func formValueAsIntOrErr(w http.ResponseWriter, r *http.Request, key string) (string, bool) {

	formVal := r.FormValue(key)

	if formVal == "" {
		errorHandler(w, r, fmt.Sprintf("http: incorrect form value %s on page %v", key, r.URL.Path))
		return "0", false
	}

	_, err := strconv.Atoi(formVal)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("http: incorrect form value %v, should be numeric - on page %v", formVal, r.URL.Path))
		return "", false
	}

	return formVal, true

}

// eg. 2026/03/19/86d276d2b8970e96.jpg
func generatePath(ext string) (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	date := time.Now().UTC().Format("2006/01/02")

	return fmt.Sprintf("%s/%s%s", date, hex.EncodeToString(b), ext), nil
}

type Customer struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type Location struct {
	Id         string `db:"id"`
	Name       string `db:"name"`
	CustomerId string `db:"customer_id"`
}

func getHomepageData(db *sqlx.DB, w http.ResponseWriter, r *http.Request) (bool, []Customer, []Location) {

	var customers []Customer
	var locations []Location

	selectData := func(data any, name string, sql string) bool {
		if err := db.SelectContext(r.Context(), data, sql); err != nil {
			errorHandler(w, r, fmt.Sprintf("db: error selecting from table '%s' - %v", name, err))
			return false
		}
		return true
	}

	ok := selectData(&customers, "customers", SelectCustomersSql)
	if !ok {
		return false, customers, locations
	}

	ok = selectData(&locations, "locations", SelectLocationsSql)
	if !ok {
		return false, customers, locations
	}

	return true, customers, locations

}

func getLocation(ctx context.Context, db *sqlx.DB, locationId, customerId int) (Location, error) {
	var location Location

	err := db.GetContext(
		ctx,
		&location,
		"SELECT * FROM location WHERE id = ? AND customer_id = ?",
		locationId,
		customerId,
	)

	return location, err
}

func LogInfo(msg string) { log.Println("INFO: " + msg) }

func LogError(msg string) { log.Println("ERROR: " + msg) }

func logVisitData(db *sqlx.DB, r *http.Request, uploadsDir string) (visitId int64, err error) {

	notes := r.FormValue("visit_notes")
	locationId := r.FormValue("location_id")

	locationInt, err := strconv.Atoi(locationId)
	if err != nil {
		return 0, fmt.Errorf("http: location value looks wrong - %v", locationInt)
	}

	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(InsertVisitSql, locationId, 1, notes)
	if err != nil {
		return 0, fmt.Errorf("sql: insert visit failed - %w", err)
	}

	vid, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if r.MultipartForm != nil {
		photos := r.MultipartForm.File["visit_photos"]

		for _, fh := range photos {
			file, err := fh.Open()
			if err != nil {
				return 0, fmt.Errorf("multipart: cannot open %s - %w", fh.Filename, err)
			}

			mimetype, ext, err := validateUpload(file)
			if err != nil {
				file.Close()
				return 0, fmt.Errorf("validateUpload: %s - %w", fh.Filename, err)
			}

			_, err = file.Seek(0, 0)
			if err != nil {
				file.Close()
				return 0, err
			}

			relPath, err := generatePath(ext)
			if err != nil {
				file.Close()
				return 0, err
			}

			err = saveThumbnail(file, relPath, uploadsDir)
			if err != nil {
				file.Close()
				return 0, err
			}

			err = saveToDisk(file, relPath, uploadsDir)
			file.Close()
			if err != nil {
				return 0, fmt.Errorf("saveToDisk: %w", err)
			}

			_, err = tx.ExecContext(
				r.Context(),
				SaveImageSql,
				vid,
				relPath,
				fh.Filename,
				mimetype,
				fh.Size,
			)
			if err != nil {
				return 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return vid, nil
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

func pathValueAsIntOrErr(w http.ResponseWriter, r *http.Request, key string) (string, bool) {

	formVal := r.PathValue(key)

	if formVal == "" {
		errorHandler(w, r, fmt.Sprintf("http: incorrect path value %s on page %v", key, r.URL.Path))
		return "0", false
	}

	_, err := strconv.Atoi(formVal)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("http: incorrect path value %v, should be numeric - on page %v", formVal, r.URL.Path))
		return "0", false
	}

	return formVal, true

}

func saveThumbnail(src io.Reader, relPath, uploadsDir string) error {
	img, _, err := image.Decode(src)
	if err != nil {
		return err
	}

	bounds := img.Bounds()

	const thumbWidth = 300
	thumbHeight := bounds.Dy() * thumbWidth / bounds.Dx()

	dstImg := image.NewRGBA(image.Rect(0, 0, thumbWidth, thumbHeight))

	draw.CatmullRom.Scale(
		dstImg,
		dstImg.Bounds(),
		img,
		bounds,
		draw.Over,
		nil,
	)

	thumbPath := filepath.Join(uploadsDir, "thumbs", relPath)

	err = os.MkdirAll(filepath.Dir(thumbPath), 0755)
	if err != nil {
		return err
	}

	tmpPath := thumbPath + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	err = jpeg.Encode(f, dstImg, &jpeg.Options{Quality: 85})
	closeErr := f.Close()

	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	if closeErr != nil {
		os.Remove(tmpPath)
		return closeErr
	}

	return os.Rename(tmpPath, thumbPath)
}

// creates the directories  before it creates the file and writes to disk
func saveToDisk(src io.Reader, relPath, uploadsDir string) error {
	fullPath := filepath.Join(uploadsDir, relPath)

	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return err
	}

	tmpPath := fullPath + ".tmp"

	dst, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	closeErr := dst.Close()

	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	if closeErr != nil {
		os.Remove(tmpPath)
		return closeErr
	}

	err = os.Rename(tmpPath, fullPath)
	if err != nil {
		return fmt.Errorf("this is the error: %v", err)
	}

	return nil
}

func setAriaValidity(val bool) string {
	if val {
		return "true"
	}
	return "false"
}

const (

	// --------------------------------------

	SaveImageSql string = `INSERT INTO images (visit_id, filename, original_name, mimetype, size, created_at) VALUES($1, $2, $3, $4, $5, CURRENT_TIMESTAMP);`

	// --------------------------------------

	InsertVisitSql string = `INSERT INTO visits (location_id, employee_id, notes) VALUES ($1, $2, $3);`

	// --------------------------------------

	SelectCustomersSql string = `SELECT * FROM customer;`

	// --------------------------------------

	SelectLocationsSql string = `SELECT * FROM location;`

	// ----------------------------------------

	SelectCustomerByIdSql string = `SELECT * FROM customer WHERE id = $1`

	// --------------------------------------

	SelectVisitDataSql string = `
			SELECT c.name customer_name, l.name location_name, e.name employee_name
			FROM visits v
         		INNER JOIN location l ON v.location_id = l.id
         		INNER JOIN employee e ON e.id = v.employee_id
         		INNER JOIN customer c ON c.id = l.customer_id
			WHERE v.id = $1`

	SelectLocationById string = `
 		SELECT
			l.name AS location_name,
			c.name AS customer_name,
			c.Id AS customer_id,
			l.id AS location_id
		FROM location l
		INNER JOIN customer c
		ON l.customer_id = c.id
		WHERE l.id = $1;`
	//----------------------------------

)
