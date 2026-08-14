package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/heyubaidullah/waqti/internal/db"
)

// removeOldLogoFile best-effort deletes the previously stored logo file,
// if any. Failure to delete isn't fatal — an orphaned file in uploadsDir
// is harmless, unlike failing to save the new logo.
func (d *Deps) removeOldLogoFile() {
	old, err := db.GetSetting(d.DB, SettingLogoURL)
	if err != nil || old == "" {
		return
	}
	_ = os.Remove(filepath.Join(d.Cfg.UploadsDir, filepath.Base(old)))
}

func (d *Deps) handleUploadLogo(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+1<<20)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		respondError(w, http.StatusBadRequest, "invalid multipart form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	filename, err := saveUpload(d.Cfg.UploadsDir, file, header)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	d.removeOldLogoFile()

	logoURL := "/uploads/" + filename
	if err := db.SetSetting(d.DB, SettingLogoURL, logoURL); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save logo")
		return
	}

	d.afterAdminWrite("logo")
	respondJSON(w, http.StatusOK, map[string]string{"logo_url": logoURL})
}

func (d *Deps) handleDeleteLogo(w http.ResponseWriter, r *http.Request) {
	d.removeOldLogoFile()

	if err := db.SetSetting(d.DB, SettingLogoURL, ""); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to remove logo")
		return
	}

	d.afterAdminWrite("logo")
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
