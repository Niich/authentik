# Generated by Django 3.1.7 on 2021-03-24 07:36

import django.db.models.deletion
from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ("authentik_flows", "0016_auto_20201202_1307"),
        ("authentik_sources_saml", "0010_samlsource_pre_authentication_flow"),
    ]

    operations = [
        migrations.AlterField(
            model_name="samlsource",
            name="pre_authentication_flow",
            field=models.ForeignKey(
                help_text="Flow used before authentication.",
                on_delete=django.db.models.deletion.CASCADE,
                related_name="source_pre_authentication",
                to="authentik_flows.flow",
            ),
        ),
    ]
