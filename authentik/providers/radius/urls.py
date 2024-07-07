"""API URLs"""

from authentik.providers.radius.api import RadiusOutpostConfigViewSet, RadiusProviderViewSet
from authentik.providers.radius.property_mappings import RadiusPropertyMappingViewSet

api_urlpatterns = [
    ("outposts/radius", RadiusOutpostConfigViewSet, "radiusprovideroutpost"),
    ("providers/radius", RadiusProviderViewSet),
    ("propertymappings/radius", RadiusPropertyMappingViewSet),
]
