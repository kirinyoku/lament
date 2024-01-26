package del

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	resp "github.com/kirinyoku/lament/internal/lib/api/response"
	"github.com/kirinyoku/lament/internal/lib/logger/sl"
	"github.com/kirinyoku/lament/internal/storage"
	"log/slog"
	"net/http"
)

type Request struct {
	Alias string `json:"alias" validate:"required"`
}

type Response struct {
	resp.Response
}

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, resp.ValidateError(validateErr))

			return
		}

		err = urlDeleter.DeleteURL(req.Alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("url not found", slog.String("alias", req.Alias))
				render.JSON(w, r, resp.Error("url not found"))

				return
			}

			log.Error("failed to delete url", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to delete url"))

			return
		}

		log.Info("url deleted", slog.String("alias", req.Alias))

		render.JSON(w, r, Response{
			Response: resp.OK(),
		})
	}
}
