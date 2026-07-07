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
	"os"
	"path/filepath"
	"strconv"

	"github.com/mugtree/grounds/app/db"

	"io"
	"net/http"
	"time"

	"github.com/goforj/godump"
	"golang.org/x/image/draw"
)

func groundsGetCustomersAndLocations(queries *db.Queries, ctx context.Context) ([]db.Customer, []db.Location, error) {

	customers, err := queries.ListCustomers(ctx)
	if err != nil {
		return []db.Customer{}, []db.Location{}, err
	}

	locations, err := queries.ListLocations(ctx)
	if err != nil {
		return []db.Customer{}, []db.Location{}, err
	}

	return customers, locations, nil

}

func groundsLogVisitData(queries *db.Queries, dbHandle *sql.DB, r *http.Request, uploadsDir string) (visitId int64, err error) {

	godump.Dump(r.Form)
	notes := r.FormValue("visit_notes")
	locationId := r.FormValue("location_id")
	visitDuration := r.FormValue("visit_duration")
	visitDate := r.FormValue("visit_date")
	visitTime := r.FormValue("visit_time")

	dur, err := strconv.ParseInt(visitDuration, 10, 64) //Atoi(visitDuration)
	if err != nil {
		return 0, fmt.Errorf("http: duration value looks wrong - %v", dur)
	}

	timeInput := visitDate + "T" + visitTime + ":00Z"

	parsedDate, err := time.Parse(time.RFC3339, timeInput)
	if err != nil {
		return 0, fmt.Errorf("validation: time parts dont form a correct date %s", parsedDate.String())
	}

	locationInt, err := strconv.Atoi(locationId)
	if err != nil {
		return 0, fmt.Errorf("http: location value looks wrong - %v", locationInt)
	}

	tx, err := dbHandle.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	qtx := queries.WithTx(tx)

	vid, err := qtx.CreateVisit(r.Context(),
		db.CreateVisitParams{
			LocationID: int64(locationInt),
			EmployeeID: 1,
			Notes:      sql.NullString{String: notes},
			Datetime:   parsedDate.Format(time.RFC3339),
			Duration:   dur,
		},
	)

	//res, err := tx.Exec(InsertVisitSql, locationId, 1, notes, parsedDate.String(), dur)
	if err != nil {
		return 0, fmt.Errorf("sql: insert visit failed - %w", err)
	}

	// vid, err := res.LastInsertId()
	// if err != nil {
	// 	return 0, err
	// }

	if r.MultipartForm != nil {
		photos := r.MultipartForm.File["visit_photos"]

		for _, fh := range photos {
			file, err := fh.Open()
			if err != nil {
				return 0, fmt.Errorf("multipart: cannot open %s - %w", fh.Filename, err)
			}

			mimetype, ext, err := groundsValidateUpload(file)
			if err != nil {
				file.Close()
				return 0, fmt.Errorf("validateUpload: %s - %w", fh.Filename, err)
			}

			_, err = file.Seek(0, 0)
			if err != nil {
				file.Close()
				return 0, err
			}

			relPath, err := _generatePath(ext)
			if err != nil {
				file.Close()
				return 0, err
			}

			err = _saveThumbnail(file, relPath, uploadsDir)
			if err != nil {
				file.Close()
				return 0, err
			}

			err = _saveToDisk(file, relPath, uploadsDir)
			file.Close()
			if err != nil {
				return 0, fmt.Errorf("saveToDisk: %w", err)
			}

			err = qtx.CreateImage(r.Context(), db.CreateImageParams{VisitID: vid, Filename: relPath, OriginalName: fh.Filename, Mimetype: sql.NullString{String: mimetype}, Size: sql.NullInt64{Int64: fh.Size}})
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

// reliaistically would likely validate more
func groundsValidateVisit(r *http.Request) (VisitTemplateData, error) {

	godump.Dump(r.Form)

	vm := VisitTemplateData{}

	// cid, err := strconv.ParseInt(r.FormValue("customer_id"), 10, 64)
	// if err != nil {
	// 	return vm, err
	// }
	// vm.CustomerId = cid

	lid, err := strconv.ParseInt(r.FormValue("location_id"), 10, 64)
	if err != nil {
		return vm, err
	}
	vm.LocationId = lid

	vm.CustomerName = r.FormValue("customer_name")
	vm.LocationName = r.FormValue("location_name")
	vm.Date = r.FormValue("visit_date")
	vm.Time = r.FormValue("visit_date")
	vm.Duration = r.FormValue("visit_duration")
	vm.IsSubmission = true

	if _hasDateError(r.FormValue("visit_date")) {
		vm.HasDateError = true
	}

	if _hasTimeError(r.FormValue("visit_time")) {
		vm.HasTimeError = true
	}

	vm.HasNotesError = false

	return vm, nil

}

// image.DecodeConfig proves that the file bytes are decodable as an image format the go actually understands
// only allows jpges atm
func groundsValidateUpload(file io.ReadSeeker) (string, string, error) {
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

type VisitTemplateData struct {
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

func (v VisitTemplateData) HasErrors() bool {
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

// eg. 2026/03/19/86d276d2b8970e96.jpg
func _generatePath(ext string) (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	date := time.Now().UTC().Format("2006/01/02")

	return fmt.Sprintf("%s/%s%s", date, hex.EncodeToString(b), ext), nil
}

func _saveThumbnail(src io.Reader, relPath, uploadsDir string) error {
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
func _saveToDisk(src io.Reader, relPath, uploadsDir string) error {
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

func _hasDateError(input string) bool {
	_, err := time.Parse("2006-01-02", input)
	if err != nil {
		return true
	}
	return false
}

func _hasTimeError(input string) bool {
	_, err := time.Parse("15:04", input)
	if err != nil {
		return true
	}
	return false
}

func _hasNotesError(_ string) bool {
	return false
}
