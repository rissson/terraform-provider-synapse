package synapse

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/lama-corp/infra/packages/gosynapse"
	"maunium.net/go/mautrix"
	mautrix_id "maunium.net/go/mautrix/id"
)

func schemaMembership() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"user_id": &schema.Schema{
			Description: "The ID of the user, format: @user_id:homeserver.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"room_id": &schema.Schema{
			Description: "The ID or alias of the room, format: !room_id:homeserver or #room_alias:homeserver. Must be a room ID for deletion to work.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
	}
}

func resourceMembership() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource can be used to manage Synapse users room memberships. The provider user must be admin in the rooms for this resource to properly work.",
		CreateContext: resourceMembershipCreate,
		ReadContext:   resourceMembershipRead,
		DeleteContext: resourceMembershipDelete,
		Schema:        schemaMembership(),
	}
}

func resourceMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*gosynapse.Client)
	var diags diag.Diagnostics

	userID := d.Get("user_id").(string)
	roomID := d.Get("room_id").(string)

	err := cli.JoinRoom(userID, roomID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s-%s", userID, roomID))

	return diags
}

func resourceMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*gosynapse.Client)
	var diags diag.Diagnostics

	cli.Cli.KickUser(mautrix_id.RoomID(d.Get("room_id").(string)), &mautrix.ReqKickUser{
		UserID: mautrix_id.UserID(d.Get("user_id").(string)),
	})

	return diags
}
