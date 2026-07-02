package profiles

import (
	"encoding/json"
	"io"
	"net/http"

	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
	"encore.dev/types/uuid"
)

// avatars stores user avatar images. It is public so the returned URLs can be
// rendered directly by browsers without authentication.
var avatars = objects.NewBucket("avatars", objects.BucketConfig{
	Public: true,
})

type uploadAvatarResponse struct {
	URL string `json:"url"`
}

// UploadAvatar stores the request body as a new avatar image and returns its
// public URL. The caller is expected to persist the URL via the profile
// endpoints. Each upload gets a fresh key, so re-uploading never overwrites.
//
//encore:api auth raw method=POST path=/profiles/avatar
func UploadAvatar(w http.ResponseWriter, req *http.Request) {
	key, err := uuid.NewV4()
	if err != nil {
		errs.HTTPError(w, errs.WrapCode(err, errs.Internal, "failed to generate avatar key"))
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	writer := avatars.Upload(req.Context(), key.String(),
		objects.WithUploadAttrs(objects.UploadAttrs{ContentType: contentType}))
	if _, err := io.Copy(writer, req.Body); err != nil {
		writer.Abort(err)
		errs.HTTPError(w, errs.WrapCode(err, errs.Internal, "failed to upload avatar"))
		return
	}
	if err := writer.Close(); err != nil {
		errs.HTTPError(w, errs.WrapCode(err, errs.Internal, "failed to store avatar"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(uploadAvatarResponse{URL: avatars.PublicURL(key.String()).String()})
}
