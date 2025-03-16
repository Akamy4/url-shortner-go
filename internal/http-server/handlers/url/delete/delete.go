package delete

import (
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/storage"
)

type Response struct {
	response.Response
	Message string `json:"message"`
}

type DeleteURL interface {
	DeleteUrl(alias string) error
}

func New(log *slog.Logger, deleteUrl DeleteURL) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.url.delete.New"

		slog.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		alias := chi.URLParam(request, "alias")

		if alias == "" {
			log.Info("alias is empty", "alias", alias)

			render.JSON(writer, request, response.Error("alias is empty"))

			return
		}

		err := deleteUrl.DeleteUrl(alias)

		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found by alias")

			render.JSON(writer, request, response.Error("url not found by alias"))

			return
		}

		log.Info("url deleted by alias: ", slog.String("alias", alias))

		render.JSON(writer, request, Response{
			Response: response.OK(),
			Message:  "alias deleted",
		})
	}
}
