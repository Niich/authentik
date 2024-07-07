"""Radius Property mappings API Views"""

from django_filters.filters import AllValuesMultipleFilter
from django_filters.filterset import FilterSet
from drf_spectacular.types import OpenApiTypes
from drf_spectacular.utils import extend_schema_field
from rest_framework.viewsets import ModelViewSet

from authentik.core.api.property_mappings import PropertyMappingSerializer
from authentik.core.api.used_by import UsedByMixin
from authentik.providers.radius.models import RadiusPropertyMapping


class RadiusPropertyMappingSerializer(PropertyMappingSerializer):
    """RadiusPropertyMapping Serializer"""

    class Meta:
        model = RadiusPropertyMapping
        fields = PropertyMappingSerializer.Meta.fields + [
            "radius_reply_attribute",
            "attibute_friendly_name",
        ]


class RadiusPropertyMappingFilter(FilterSet):
    """Filter for RadiusPropertyMapping"""

    managed = extend_schema_field(OpenApiTypes.STR)(AllValuesMultipleFilter(field_name="managed"))

    class Meta:
        model = RadiusPropertyMapping
        fields = "__all__"


class RadiusPropertyMappingViewSet(UsedByMixin, ModelViewSet):
    """RadiusPropertyMapping Viewset"""

    queryset = RadiusPropertyMapping.objects.all()
    serializer_class = RadiusPropertyMappingSerializer
    filterset_class = RadiusPropertyMappingFilter
    search_fields = ["name"]
    ordering = ["name"]
