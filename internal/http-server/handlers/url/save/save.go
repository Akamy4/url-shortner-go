package save

import (
	"errors"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

const AliasLength = 6

type URLServer interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlServer URLServer) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.url.save.New"

		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		var req Request

		err := render.DecodeJSON(request.Body, &req)

		if err != nil {
			logger.Error("failed to decode request body", sl.Error(err))

			render.JSON(writer, request, response.Error("failed to decode request"))

			return
		}

		logger.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			logger.Error("invalid request", sl.Error(err))

			render.JSON(writer, request, response.ValidationError(validateErr))

			return
		}

		alias := req.Alias

		if alias == "" {
			alias = random.NewRandomString(AliasLength)
		}

		id, err := urlServer.SaveURL(req.URL, alias)

		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(writer, request, response.Error("url already exists"))

			return
		}

		if err != nil {
			log.Error("failed to add url", sl.Error(err))

			render.JSON(writer, request, response.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		render.JSON(writer, request, Response{
			Response: response.OK(),
			Alias:    alias,
		})

	}
}
