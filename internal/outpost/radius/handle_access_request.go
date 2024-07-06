package radius

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"goauthentik.io/internal/outpost/flow"
	"goauthentik.io/internal/outpost/radius/metrics"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func (rs *RadiusServer) Handle_AccessRequest(w radius.ResponseWriter, r *RadiusRequest) {
	username := rfc2865.UserName_GetString(r.Packet)

	fe := flow.NewFlowExecutor(r.Context(), r.pi.flowSlug, r.pi.s.ac.Client.GetConfig(), log.Fields{
		"username":  username,
		"client":    r.RemoteAddr(),
		"requestId": r.ID(),
	})
	fe.DelegateClientIP(r.RemoteAddr())
	fe.Params.Add("goauthentik.io/outpost/radius", "true")

	fe.Answers[flow.StageIdentification] = username
	fe.SetSecrets(rfc2865.UserPassword_GetString(r.Packet), r.pi.MFASupport)

	passed, err := fe.Execute()
	if err != nil {
		r.Log().WithField("username", username).WithError(err).Warning("failed to execute flow")
		metrics.RequestsRejected.With(prometheus.Labels{
			"outpost_name": rs.ac.Outpost.Name,
			"reason":       "flow_error",
			"app":          r.pi.appSlug,
		}).Inc()
		_ = w.Write(r.Response(radius.CodeAccessReject))
		return
	}
	if !passed {
		metrics.RequestsRejected.With(prometheus.Labels{
			"outpost_name": rs.ac.Outpost.Name,
			"reason":       "invalid_credentials",
			"app":          r.pi.appSlug,
		}).Inc()
		_ = w.Write(r.Response(radius.CodeAccessReject))
		return
	}
	access, err := fe.CheckApplicationAccess(r.pi.appSlug)
	if err != nil {
		r.Log().WithField("username", username).WithError(err).Warning("failed to check access")
		_ = w.Write(r.Response(radius.CodeAccessReject))
		metrics.RequestsRejected.With(prometheus.Labels{
			"outpost_name": rs.ac.Outpost.Name,
			"reason":       "access_check_fail",
			"app":          r.pi.appSlug,
		}).Inc()
		return
	}
	if !access {
		r.Log().WithField("username", username).Info("Access denied for user")
		_ = w.Write(r.Response(radius.CodeAccessReject))
		metrics.RequestsRejected.With(prometheus.Labels{
			"outpost_name": rs.ac.Outpost.Name,
			"reason":       "access_denied",
			"app":          r.pi.appSlug,
		}).Inc()
		return
	}

	// Now that we know they have access we can add attributes to the response

	// create the response packet
	responsePacket := r.Response(radius.CodeAccessAccept)

	// use the akAPI to get attribute values to add from the application
	// rs.ac.Client.PropertymappingsApi.PropertymappingsSamlList()
	// r.pi.s.

	attr, err := radius.NewString("shell:priv-lvl=15")
	if err != nil {
		r.Log().WithField("username", username).WithError(err).Warning("failed to create attribute")
		metrics.RequestsRejected.With(prometheus.Labels{
			"outpost_name": rs.ac.Outpost.Name,
			"reason":       "attribute_create_fail",
			"app":          r.pi.appSlug,
		}).Inc()
		return
	}
	r.Log().WithField("username", username).WithField("before", len(responsePacket.Attributes)).Debug(responsePacket.Attributes)

	err = r.AddVendor_Attribute(responsePacket, 9, 1, attr)
	r.Log().WithField("username", username).WithField("after", len(responsePacket.Attributes)).Debug(responsePacket.Attributes)
	if err != nil {
		r.Log().WithField("username", username).WithError(err).Warning("failed to create attribute")
		metrics.RequestsRejected.With(prometheus.Labels{
			"outpost_name": rs.ac.Outpost.Name,
			"reason":       "attribute_create_fail",
			"app":          r.pi.appSlug,
		}).Inc()
		return
	}

	_ = w.Write(responsePacket)
}
