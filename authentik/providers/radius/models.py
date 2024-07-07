"""Radius Provider"""

from django.db import models
from django.templatetags.static import static
from django.utils.translation import gettext_lazy as _
from rest_framework.serializers import Serializer

from authentik.core.models import PropertyMapping, Provider
from authentik.lib.generators import generate_id
from authentik.outposts.models import OutpostModel


class RadiusProvider(OutpostModel, Provider):
    """Allow applications to authenticate against authentik's users using Radius."""

    shared_secret = models.TextField(
        default=generate_id,
        help_text=_("Shared secret between clients and server to hash packets."),
    )

    client_networks = models.TextField(
        default="0.0.0.0/0, ::/0",
        help_text=_(
            "List of CIDRs (comma-separated) that clients can connect from. A more specific "
            "CIDR will match before a looser one. Clients connecting from a non-specified CIDR "
            "will be dropped."
        ),
    )

    mfa_support = models.BooleanField(
        default=True,
        verbose_name="MFA Support",
        help_text=_(
            "When enabled, code-based multi-factor authentication can be used by appending a "
            "semicolon and the TOTP code to the password. This should only be enabled if all "
            "users that will bind to this provider have a TOTP device configured, as otherwise "
            "a password may incorrectly be rejected if it contains a semicolon."
        ),
    )

    @property
    def launch_url(self) -> str | None:
        """Radius never has a launch URL"""
        return None

    @property
    def component(self) -> str:
        return "ak-provider-radius-form"

    @property
    def icon_url(self) -> str | None:
        return static("authentik/sources/radius.svg")

    @property
    def serializer(self) -> type[Serializer]:
        from authentik.providers.radius.api import RadiusProviderSerializer

        return RadiusProviderSerializer

    def __str__(self):
        return f"Radius Provider {self.name}"

    class Meta:
        verbose_name = _("Radius Provider")
        verbose_name_plural = _("Radius Providers")

class RadiusPropertyMapping(PropertyMapping):
    """Map User/Group attribute to RADIUS Attribute, which can be used by the Service Provider"""

    radius_reply_attribute = models.TextField(verbose_name="RADIUS Name")
    attibute_friendly_name = models.TextField(default=None, blank=True, null=True)

    @property
    def component(self) -> str:
        return "ak-property-mapping-radius-form"

    @property
    def serializer(self) -> type[Serializer]:
        from authentik.providers.radius.property_mappings import RadiusPropertyMappingSerializer

        return RadiusPropertyMappingSerializer

    def __str__(self):
        name = self.attibute_friendly_name if self.attibute_friendly_name != "" else self.radius_reply_attribute
        return f"{self.name} ({name})"

    class Meta:
        verbose_name = _("RADIUS Property Mapping")
        verbose_name_plural = _("RADIUS Property Mappings")
