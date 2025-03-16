package redirect

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

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		alias := chi.URLParam(request, "alias")

		if alias == "" {
			log.Info("alias is empty", "alias", alias)

			render.JSON(writer, request, response.Error("alias is empty"))

			return
		}

		url, err := urlGetter.GetURL(alias)

		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found by alias")

			render.JSON(writer, request, response.Error("url not found by alias"))

			return
		}

		log.Info("got url", slog.String("url", url))

		// redirect to url
		http.Redirect(writer, request, url, http.StatusFound)
	}
}
