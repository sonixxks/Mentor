// Package profilehttp — REST-ручки анкеты: /profiles/me и материалы для верификации.
// Отдельный пакет, потому что middleware зависит от auth, а auth — от profile.
package profilehttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/platform/authctx"
	"github.com/sonixxks/Mentor/backend/internal/profile"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/httpx"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/middleware"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type ProfileService interface {
	Get(ctx context.Context, userID uuid.UUID) (profile.Profile, error)
	Patch(ctx context.Context, userID uuid.UUID, raw map[string]json.RawMessage) (profile.Profile, error)
	AddAssets(ctx context.Context, actor uuid.UUID, role user.Role, files []profile.UploadedFile, links []string) (profile.Profile, error)
	DeleteAsset(ctx context.Context, actor uuid.UUID, role user.Role, assetID uuid.UUID) (profile.Profile, error)
}

type Handler struct {
	svc  ProfileService
	auth *middleware.Auth
	log  *slog.Logger
}

func NewHandler(svc ProfileService, auth *middleware.Auth, log *slog.Logger) *Handler {
	return &Handler{svc: svc, auth: auth, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /profiles/me", h.auth.Require(h.get))
	mux.Handle("PATCH /profiles/me", h.auth.Require(h.patch))
	mux.Handle("POST /profiles/me/verification-assets", h.auth.Require(h.addAssets))
	mux.Handle("DELETE /profiles/me/verification-assets/{assetId}", h.auth.Require(h.deleteAsset))
}

// profileJSON — ProfileDetailResponse. Незаполненные поля приходят как null, чтобы форме было проще.
type profileJSON struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Role   user.Role `json:"role"`

	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName *string `json:"middle_name"`
	AvatarURL  *string `json:"avatar_url"`

	Age           *int    `json:"age"`
	CurrentStatus *string `json:"current_status"`
	University    *string `json:"university"`
	Major         *string `json:"major"`
	Course        *int    `json:"course"`
	Experience    *string `json:"experience"`

	JobTitle    *string  `json:"job_title"`
	Company     *string  `json:"company"`
	Bio         *string  `json:"bio"`
	HourlyRate  *float64 `json:"hourly_rate"`
	MaxStudents *int     `json:"max_students"`
	IsPublished bool     `json:"is_published"`

	CommunicationStyle *string          `json:"communication_style"`
	AvailabilitySlots  map[string][]int `json:"availability_slots"`
	MainGoals          []string         `json:"main_goals"`
	SuccessMetric      *string          `json:"success_metric"`
	TimeLimit          *int             `json:"time_limit"`
	NeedsCuratorHelp   bool             `json:"needs_curator_help"`
	MinDurationMonths  *int             `json:"min_duration_months"`
	MaxDurationMonths  *int             `json:"max_duration_months"`

	CurrentSkills           []profile.Skill             `json:"current_skills"`
	TargetSkills            []string                    `json:"target_skills"`
	HelpTopics              []string                    `json:"help_topics"`
	TargetStudentLevels     []string                    `json:"target_student_levels"`
	NotificationPreferences map[string]profile.Channels `json:"notification_preferences"`

	OnboardingCompleted bool            `json:"onboarding_completed"`
	VerificationStatus  string          `json:"verification_status"`
	VerificationAssets  []profile.Asset `json:"verification_assets"`
	FreeSlots           *int            `json:"free_slots,omitempty"`
}

type verificationJSON struct {
	VerificationStatus string          `json:"verification_status"`
	VerificationAssets []profile.Asset `json:"verification_assets"`
}

func toJSON(p profile.Profile) profileJSON {
	out := profileJSON(p)
	out.AvailabilitySlots = orEmptyMap(out.AvailabilitySlots)
	out.MainGoals = orEmpty(out.MainGoals)
	out.CurrentSkills = orEmpty(out.CurrentSkills)
	out.TargetSkills = orEmpty(out.TargetSkills)
	out.HelpTopics = orEmpty(out.HelpTopics)
	out.TargetStudentLevels = orEmpty(out.TargetStudentLevels)
	out.VerificationAssets = orEmpty(out.VerificationAssets)
	return out
}

func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func orEmptyMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return map[K]V{}
	}
	return m
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.Get(r.Context(), principal(r).UserID)
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusOK, toJSON(p))
}

func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	var raw map[string]json.RawMessage
	if !httpx.DecodeJSON(w, r, &raw, false) {
		return
	}
	p, err := h.svc.Patch(r.Context(), principal(r).UserID, raw)
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusOK, toJSON(p))
}

const maxMultipartMemory = 8 << 20

func (h *Handler) addAssets(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, profile.MaxFilesPerCall*profile.MaxAssetFileSize+(1<<20))
	if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
		httpx.SendBadRequest(w, "ожидается multipart/form-data с полями files и links (не больше 5 файлов по 10 МБ)")
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	var files []profile.UploadedFile
	for _, fh := range r.MultipartForm.File["files"] {
		files = append(files, profile.UploadedFile{Name: fh.Filename, Size: fh.Size, Open: openPart(fh)})
	}

	p := principal(r)
	profile, err := h.svc.AddAssets(r.Context(), p.UserID, p.Role, files, r.MultipartForm.Value["links"])
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusOK, verificationJSON{profile.VerificationStatus, orEmpty(profile.VerificationAssets)})
}

func (h *Handler) deleteAsset(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "assetId")
	if err != nil {
		httpx.SendError(w, http.StatusNotFound, httpx.ErrorResponse{ErrorCode: httpx.ErrNotFound, Message: profile.ErrAssetNotFound.Error()})
		return
	}
	p := principal(r)
	profile, err := h.svc.DeleteAsset(r.Context(), p.UserID, p.Role, id)
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusOK, verificationJSON{profile.VerificationStatus, orEmpty(profile.VerificationAssets)})
}

func (h *Handler) sendError(w http.ResponseWriter, err error) {
	if v, ok := profile.IsValidation(err); ok {
		httpx.SendBadRequest(w, httpx.MsgValidation, v.Details...)
		return
	}
	switch {
	case errors.Is(err, profile.ErrProfileNotFound), errors.Is(err, profile.ErrAssetNotFound):
		httpx.SendError(w, http.StatusNotFound, httpx.ErrorResponse{ErrorCode: httpx.ErrNotFound, Message: err.Error()})
	case errors.Is(err, profile.ErrMentorOnly):
		httpx.SendError(w, http.StatusForbidden, httpx.ErrorResponse{ErrorCode: httpx.ErrForbidden, Message: err.Error()})
	default:
		h.log.Error("profiles handler", "error", err)
		httpx.SendInternalError(w)
	}
}

func openPart(fh *multipart.FileHeader) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) { return fh.Open() }
}

func principal(r *http.Request) authctx.Principal {
	p, _ := middleware.PrincipalFrom(r.Context())
	return p
}
