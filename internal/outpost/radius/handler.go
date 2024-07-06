package radius

import (
	"crypto/sha512"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"goauthentik.io/internal/outpost/radius/metrics"
	"goauthentik.io/internal/utils"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type RadiusRequest struct {
	*radius.Request
	log  *log.Entry
	id   string
	span *sentry.Span
	pi   *ProviderInstance
}

func (r *RadiusRequest) Log() *log.Entry {
	return r.log
}

func (r *RadiusRequest) RemoteAddr() string {
	return utils.GetIP(r.Request.RemoteAddr)
}

func (r *RadiusRequest) ID() string {
	return r.id
}

func (rs *RadiusServer) ServeRADIUS(w radius.ResponseWriter, r *radius.Request) {
	span := sentry.StartSpan(r.Context(), "authentik.providers.radius.connect",
		sentry.WithTransactionName("authentik.providers.radius.connect"))
	rid := uuid.New().String()
	span.SetTag("request_uid", rid)
	rl := rs.log.WithField("code", r.Code.String()).WithField("request", rid)
	selectedApp := ""
	defer func() {
		span.Finish()
		metrics.Requests.With(prometheus.Labels{
			"outpost_name": rs.ac.Outpost.Name,
			"app":          selectedApp,
		}).Observe(float64(span.EndTime.Sub(span.StartTime)) / float64(time.Second))
	}()

	nr := &RadiusRequest{
		Request: r,
		log:     rl,
		id:      rid,
		span:    span,
	}

	rl.Info("Radius Request")

	// Lookup provider by shared secret
	var pi *ProviderInstance
	for _, p := range rs.providers {
		if string(p.SharedSecret) == string(r.Secret) {
			pi = p
			selectedApp = pi.appSlug
			break
		}
	}
	if pi == nil {
		nr.Log().WithField("hashed_secret", string(sha512.New().Sum(r.Secret))).Warning("No provider found")
		_ = w.Write(r.Response(radius.CodeAccessReject))
		return
	}
	nr.pi = pi

	if nr.Code == radius.CodeAccessRequest {
		rs.Handle_AccessRequest(w, nr)
	}
}

func (r *RadiusRequest) AddVendor_Attribute(p *radius.Packet, venderId uint32, typ byte, attr radius.Attribute) (err error) {
	var vsa radius.Attribute
	vendor := make(radius.Attribute, 2+len(attr))
	vendor[0] = typ
	vendor[1] = byte(len(vendor))
	copy(vendor[2:], attr)
	vsa, err = radius.NewVendorSpecific(venderId, vendor)
	if err != nil {
		return err
	}
	p.Add(rfc2865.VendorSpecific_Type, vsa)
	return
}
