package app

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
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
	"net/url"
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

func hasDateError(input string) bool {
	_, err := time.Parse("2006-01-02", input)
	if err != nil {
		return true
	}
	return false
}

func hasTimeError(input string) bool {
	_, err := time.Parse("15:04", input)
	if err != nil {
		return true
	}
	return false
}

func hasNotesError(_ string) bool {
	return false
}

func validateVisit(r *http.Request) VisitVM {

	vm := VisitVM{}
	/* in a realistic scenario we would validate
	everything
	not just date and time
	*/

	vm.CustomerId = r.FormValue("customer_id")
	vm.CustomerName = r.FormValue("customer_name")
	vm.LocationId = r.FormValue("location_id")
	vm.LocationName = r.FormValue("location_name")
	vm.Date = r.FormValue("visit_date")
	vm.Time = r.FormValue("visit_date")
	vm.Duration = r.FormValue("visit_duration")
	vm.IsSubmission = true

	if hasDateError(r.FormValue("visit_date")) {
		vm.HasDateError = true
	}

	if hasTimeError(r.FormValue("visit_time")) {
		vm.HasTimeError = true
	}

	vm.HasNotesError = false

	return vm

}

// image.DecodeConfig proves that the file bytes are decodable as an image format the go actually understands
// only allows jpges atm
func validateUpload(file io.ReadSeeker) (string, string, error) {
	buf := make([]byte, 512)

	n, err := file.Read(buf)
	if err != nil {
		return "", "", err
	}

	mimeType := http.DetectContentType(buf[:n])

	_, err = file.Seek(0, 0)
	if err != nil {
		return "", "", err
	}

	_, _, err = image.DecodeConfig(file)
	if err != nil {
		return "", "", errors.New("invalid image data")
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return "", "", err
	}

	if mimeType != "image/jpeg" {
		return "", "", errors.New("unsupported image type")
	}

	return mimeType, ".jpg", nil
}

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

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
)

func writeCookie(w http.ResponseWriter, cookie http.Cookie) error {
	// Encode the cookie value using base64.
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))

	// Check the total length of the cookie contents. Return the ErrValueTooLong
	// error if it's more than 4096 bytes.
	if len(cookie.String()) > 4096 {
		return ErrValueTooLong
	}

	// Write the cookie as normal.
	http.SetCookie(w, &cookie)

	return nil
}

func readCookie(r *http.Request, name string) (string, error) {
	// Read the cookie as normal.
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	// Decode the base64-encoded cookie value. If the cookie didn't contain a
	// valid base64-encoded value, this operation will fail and we return an
	// ErrInvalidValue error.
	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", ErrInvalidValue
	}

	// Return the decoded cookie value.
	return string(value), nil
}

func writeSignedCookie(w http.ResponseWriter, cookie http.Cookie, secretKey []byte) error {
	// Calculate a HMAC signature of the cookie name and value, using SHA256 and
	// a secret key (which we will create in a moment).
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)

	// Prepend the cookie value with the HMAC signature.
	cookie.Value = string(signature) + cookie.Value

	// Call our Write() helper to base64-encode the new cookie value and write
	// the cookie.
	return writeCookie(w, cookie)
}

func readSignedCookie(r *http.Request, name string, secretKey []byte) (string, error) {
	// Read in the signed value from the cookie. This should be in the format
	// "{signature}{original value}".
	signedValue, err := readCookie(r, name)
	if err != nil {
		return "", err
	}

	// A SHA256 HMAC signature has a fixed length of 32 bytes. To avoid a potential
	// 'index out of range' panic in the next step, we need to check sure that the
	// length of the signed cookie value is at least this long. We'll use the
	// sha256.Size constant here, rather than 32, just because it makes our code
	// a bit more understandable at a glance.
	if len(signedValue) < sha256.Size {
		return "", ErrInvalidValue
	}

	// Split apart the signature and original cookie value.
	signature := signedValue[:sha256.Size]
	value := signedValue[sha256.Size:]

	// Recalculate the HMAC signature of the cookie name and original value.
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	// Check that the recalculated signature matches the signature we received
	// in the cookie. If they match, we can be confident that the cookie name
	// and value haven't been edited by the client.
	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", ErrInvalidValue
	}

	// Return the original cookie value.
	return value, nil
}
