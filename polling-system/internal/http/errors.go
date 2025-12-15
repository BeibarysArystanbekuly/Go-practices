package api

import (
	"database/sql"
	"errors"
	"net/http"

	"polling-system/internal/domain/poll"
	"polling-system/internal/domain/user"
	"polling-system/internal/domain/vote"
	"polling-system/internal/platform/apperr"
)

func errorResponse(w http.ResponseWriter, err error) {
	appErr := mapError(err)
	writeJSON(w, appErr.StatusCode(), map[string]string{
		"error":   appErr.Code,
		"message": appErr.Message,
	})
}

func mapError(err error) *apperr.AppError {
	if err == nil {
		return apperr.Internal("internal_error", "internal server error", nil)
	}

	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return apperr.NotFound("not_found", "resource not found", err)
	case errors.Is(err, user.ErrInvalidCredentials):
		return apperr.Unauthorized("invalid_credentials", "invalid credentials", err)
	case errors.Is(err, user.ErrInactiveUser):
		return apperr.Unauthorized("inactive_user", "user is inactive", err)
	case errors.Is(err, user.ErrEmailTaken):
		return apperr.BadRequest("email_taken", "email already taken", err)
	case errors.Is(err, poll.ErrPollNotFound):
		return apperr.NotFound("poll_not_found", "poll not found", err)
	case errors.Is(err, poll.ErrInvalidStatus):
		return apperr.BadRequest("invalid_status", "invalid poll status", err)
	case errors.Is(err, poll.ErrInvalidDates):
		return apperr.BadRequest("invalid_dates", "ends_at must be after starts_at", err)
	case errors.Is(err, vote.ErrAlreadyVoted):
		return apperr.Conflict("already_voted", "user already voted in this poll", err)
	case errors.Is(err, vote.ErrPollNotActive):
		return apperr.BadRequest("poll_not_active", "poll is not active", err)
	case errors.Is(err, vote.ErrOptionNotInPoll):
		return apperr.BadRequest("invalid_option", "option does not belong to poll", err)
	case errors.Is(err, vote.ErrPollNotFound):
		return apperr.NotFound("poll_not_found", "poll not found", err)
	default:
		return apperr.Internal("internal_error", http.StatusText(http.StatusInternalServerError), err)
	}
}
