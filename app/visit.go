package app

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"strconv"

	"io"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/image/draw"
)

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
