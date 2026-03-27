package app

import (
	"context"
	"crypto/rand"
	"database/sql"
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

	"io"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/image/draw"
)

type HomepageVm struct {
	SelectedCustomer int
	SelectedLocation int
	ShowLocations    bool
	Customers        []Customer
	Locations        []Location
	IsValid          bool
}

type PickCustomerVm struct {
	Customers []Customer
	HasError  bool
	//PreviousVisits []
}

type PickLocationVm struct {
	CustomerId   int
	CustomerName string
	Locations    []Location
	HasError     bool
}

type homePageSignals struct {
	CustomerId int `json:"customerId"`
	LocationId int `json:"locationId"`
}

func formValueAsIntOrErr(w http.ResponseWriter, r *http.Request, key string) (int, bool) {

	formVal := r.FormValue(key)

	if formVal == "" {
		renderServerError(w, r, fmt.Sprintf("http: incorrect form value %s on page %v", key, r.URL.Path))
		return 0, false
	}

	val, err := strconv.Atoi(formVal)
	if err != nil {
		renderServerError(w, r, fmt.Sprintf("http: incorrect form value %v, should be numeric - on page %v", formVal, r.URL.Path))
		return 0, false
	}

	return val, true

}

func pathValueAsIntOrErr(w http.ResponseWriter, r *http.Request, key string) (int, bool) {

	formVal := r.PathValue(key)

	if formVal == "" {
		renderServerError(w, r, fmt.Sprintf("http: incorrect path value %s on page %v", key, r.URL.Path))
		return 0, false
	}

	val, err := strconv.Atoi(formVal)
	if err != nil {
		renderServerError(w, r, fmt.Sprintf("http: incorrect path value %v, should be numeric - on page %v", formVal, r.URL.Path))
		return 0, false
	}

	return val, true

}

func getHomepageData(db *sqlx.DB, w http.ResponseWriter, r *http.Request) (bool, []Customer, []Location) {

	var customers []Customer
	var locations []Location

	selectData := func(data any, name string, sql string) bool {
		if err := db.SelectContext(r.Context(), data, sql); err != nil {
			renderServerError(w, r, fmt.Sprintf("db: error selecting from table '%s' - %v", name, err))
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

func filteredLocations(locations []Location, customerId int) []Location {
	filtered := make([]Location, 0, len(locations))
	for _, loc := range locations {
		if loc.CustomerId == customerId {
			filtered = append(filtered, loc)
		}
	}
	return filtered
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

func handleLocationError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	signals homePageSignals,
) {
	if err == sql.ErrNoRows {
		renderServerError(
			w,
			r,
			fmt.Sprintf(
				"sql: error selecting location - check inputs - %v - %v",
				signals.LocationId,
				signals.CustomerId,
			),
		)
		return
	}

	renderServerError(
		w,
		r,
		fmt.Sprintf(
			"http: error selecting location - check inputs - %v - %v",
			signals.LocationId,
			signals.CustomerId,
		),
	)
}

func validateDate(input string) bool {
	_, err := time.Parse("2006-01-02", input)
	if err != nil {
		return false
	}
	return true
}

func validateTime(input string) bool {
	_, err := time.Parse("15:04", input)
	if err != nil {
		return false
	}
	return true
}

func validateVisitSubmission(r *http.Request) visitVM {

	vm := visitVM{}
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

	if r.FormValue("visit_date") == "" {
		vm.HasDateError = true
	}

	if r.FormValue("visit_time") == "" {
		vm.HasTimeError = true
	}

	return vm

}

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
		photos := r.MultipartForm.File["original-photos"]

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

	// SelectLocationsByCustomerIdSql = `
	// 	SELECT
	// 		l.name AS location_name,
	// 		c.name AS customer_name,
	// 		l.id AS location_id
	// 	FROM location l
	// 	INNER JOIN customer c
	// 	ON l.customer_id = c.id
	// 	WHERE c.id = $1;`

	// --------------------------------------

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

type Customer struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type Employee struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type Location struct {
	Id         int    `db:"id"`
	Name       string `db:"name"`
	CustomerId int    `db:"customer_id"`
}

type Visit struct {
	Id         int `db:"id"`
	EmployeeId int `db:"employee_id"`
	LocationId int `db:"location_id"`
}

type locationByCustomer struct {
	LocationName string `db:"location_name"`
	CustomerName string `db:"customer_name"`
	CustomerId   string `db:"customer_id"`
	LocationId   string `db:"location_id"`
}

type getLocSignals struct {
	CustomerId string `json:"customerId"`
}

type visitVM struct {
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

func (v visitVM) HasErrors() bool {
	if v.HasDateError || v.HasTimeError {
		return true
	}
	return false
}

type VisitVMErrors struct {
	HasTimeError bool
	HasDateError bool
}

type ConfirmationVm struct {
	LocationId string
	VisitId    string
	Time       string
	Date       string
	Duration   string
	ImagePaths []string
}

func LogInfo(msg string)  { log.Println("INFO: " + msg) }
func LogError(msg string) { log.Println("ERROR: " + msg) }
