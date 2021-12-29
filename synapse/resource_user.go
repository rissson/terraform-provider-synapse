package synapse

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/thoas/go-funk"
	"gitlab.com/lama-corp/infra/packages/gosynapse"
)

func schemaUser() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"user_id": &schema.Schema{
			Description: "The ID of the user, format: @user_id:homeserver.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"password": &schema.Schema{
			Description: "The password of the user.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
		"display_name": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"threepids": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"medium": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
					"address": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
					"added_at": &schema.Schema{
						Type:     schema.TypeString,
						Computed: true,
					},
					"validated_at": &schema.Schema{
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"external_ids": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"auth_provider": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
					"external_id": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
		"avatar_url": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"admin": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"deactivated": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"user_type": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"shadow_banned": &schema.Schema{
			Type:     schema.TypeBool,
			Computed: true,
		},
		"password_hash": &schema.Schema{
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},
		"creation_ts": &schema.Schema{
			Type:     schema.TypeInt,
			Computed: true,
		},
		"appservice_id": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
		"consent_server_notice_sent": &schema.Schema{
			Type:     schema.TypeInt,
			Computed: true,
		},
		"consent_version": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource can be used to manage Synapse users.",
		CreateContext: resourceUserCreateOrUpdate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserCreateOrUpdate,
		DeleteContext: resourceUserDelete,
		Schema:        schemaUser(),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func reqThreePIDFromSchemaThreePID(d *schema.ResourceData) gosynapse.ReqThreePID {
	return gosynapse.ReqThreePID{
		Medium:  d.Get("medium").(string),
		Address: d.Get("address").(string),
	}
}

func reqExternalIDFromSchemaExternalID(d *schema.ResourceData) gosynapse.ReqExternalID {
	return gosynapse.ReqExternalID{
		AuthProvider: d.Get("auth_provider").(string),
		ExternalID:   d.Get("external_id").(string),
	}
}

func reqUserFromSchemaUser(d *schema.ResourceData) gosynapse.ReqUser {
	user := gosynapse.ReqUser{}

	if password, ok := d.GetOk("password"); ok {
		user.Password = password.(string)
	}

	if displayName, ok := d.GetOk("display_name"); ok {
		user.DisplayName = displayName.(string)
	}

	if avatarURL, ok := d.GetOk("avatar_url"); ok {
		user.AvatarURL = avatarURL.(string)
	}

	if admin, ok := d.GetOk("admin"); ok {
		user.Admin = admin.(bool)
	}

	if deactivated, ok := d.GetOk("deactivated"); ok {
		user.Deactivated = deactivated.(bool)
	}

	if userType, ok := d.GetOk("user_type"); ok {
		user.UserType = userType.(string)
	}

	if threePIDs, ok := d.GetOk("threepids"); ok {
		for _, threePID := range threePIDs.([]*schema.ResourceData) {
			user.ThreePIDs = append(user.ThreePIDs, reqThreePIDFromSchemaThreePID(threePID))
		}
	}

	if externalIDs, ok := d.GetOk("external_ids"); ok {
		for _, externalID := range externalIDs.([]*schema.ResourceData) {
			user.ExternalIDs = append(user.ExternalIDs, reqExternalIDFromSchemaExternalID(externalID))
		}
	}

	return user
}

func flattenRespThreePID(threePID gosynapse.RespThreePID) map[string]interface{} {
	return map[string]interface{}{
		"medium":       threePID.Medium,
		"address":      threePID.Address,
		"added_at":     threePID.AddedAt,
		"validated_at": threePID.ValidatedAt,
	}
}

func flattenRespExternalID(externalID gosynapse.RespExternalID) map[string]interface{} {
	return map[string]interface{}{
		"auth_provider": externalID.AuthProvider,
		"external_id":   externalID.ExternalID,
	}
}

func flattenRespUser(user *gosynapse.RespUser) map[string]interface{} {
	var admin bool
	if user.Admin == 0 {
		admin = false
	} else {
		admin = true
	}

	var deactivated bool
	if user.Deactivated == 0 {
		deactivated = false
	} else {
		deactivated = true
	}

	return map[string]interface{}{
		"display_name":               user.DisplayName,
		"threepids":                  funk.Map(user.ThreePIDs, flattenRespThreePID),
		"external_ids":               funk.Map(user.ExternalIDs, flattenRespExternalID),
		"avatar_url":                 user.AvatarURL,
		"admin":                      admin,
		"deactivated":                deactivated,
		"shadow_banned":              user.ShadowBanned,
		"password_hash":              user.PasswordHash,
		"creation_ts":                user.CreationTs,
		"appservice_id":              user.AppserviceID,
		"consent_server_notice_sent": user.ConsentServerNoticeSent,
		"consent_version":            user.ConsentVersion,
		"user_type":                  user.UserType,
	}
}

func resourceUserCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*gosynapse.Client)
	var diags diag.Diagnostics

	userID := d.Get("user_id").(string)
	req := reqUserFromSchemaUser(d)

	err := cli.CreateOrModifyUser(userID, &req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(userID)

	diags = resourceUserRead(ctx, d, m)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*gosynapse.Client)
	var diags diag.Diagnostics

	userID, ok := d.GetOk("user_id")
	if ok {
		d.SetId(userID.(string))
	}

	user, err := cli.GetUser(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	for key, value := range flattenRespUser(user) {
		err := d.Set(key, value)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*gosynapse.Client)
	var diags diag.Diagnostics

	err := cli.DeactivateUser(d.Id(), true)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
