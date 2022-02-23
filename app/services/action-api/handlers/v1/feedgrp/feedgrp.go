package feedgrp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/data/schema"
	"github.com/jnkroeker/makulu/business/feeds/loader"
	"github.com/jnkroeker/makulu/foundation/web"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Handlers struct {
	Log          *zap.SugaredLogger
	GqlConfig    data.GraphQLConfig
	LoaderConfig loader.Config
}

func (h *Handlers) Upload(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var request schema.UploadFeedRequest
	if err := web.Decode(r, &request); err != nil {
		return errors.Wrap(err, "decoding request")
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	go func() {
		h.Log.Debug("%s: started G: %s %s -> %s",
			v.TraceID,
			r.Method, r.URL.Path, r.RemoteAddr,
		)

		search := loader.Search{
			Name: request.Name,
			Lat:  request.Lat,
			Lng:  request.Lng,
		}
		if err := loader.UpdateData(h.Log, h.GqlConfig, v.TraceID, h.LoaderConfig, search); err != nil {
			h.Log.Debug("%s: completed G: %s %s -> %s (%d) (%s) : ERROR : %v",
				v.TraceID,
				r.Method, r.URL.Path, r.RemoteAddr,
				v.StatusCode, time.Since(v.Now), err,
			)
			return
		}

		h.Log.Debug("%s: completed G: %s %s -> %s (%d) (%s)",
			v.TraceID,
			r.Method, r.URL.Path, r.RemoteAddr,
			v.StatusCode, time.Since(v.Now),
		)
	}()

	resp := schema.UploadFeedResponse{
		Name:    request.Name,
		Lat:     request.Lat,
		Lng:     request.Lng,
		Message: fmt.Sprintf("Uploading data for user %q [%f,%f]", request.Name, request.Lat, request.Lng),
	}
	return web.Respond(ctx, w, resp, http.StatusOK)
}
