package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsIdentityStoreGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIdentityStoreGroupRead,

		Schema: map[string]*schema.Schema{
			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},

			"group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"display_name"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},

			"display_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"group_id"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{S}\p{N}\p{P}\t\n\r ]+$`), "must match [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}\\t\\n\\r ]"),
				),
			},
		},
	}
}

func dataSourceAwsIdentityStoreGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).identitystoreconn

	identityStoreID := d.Get("identity_store_id").(string)
	groupID := d.Get("group_id").(string)
	displayName := d.Get("display_name").(string)

	if groupID != "" {
		log.Printf("[DEBUG] Reading AWS Identity Store Group")
		resp, err := conn.DescribeGroup(&identitystore.DescribeGroupInput{
			IdentityStoreId: aws.String(identityStoreID),
			GroupId:         aws.String(groupID),
		})
		if err != nil {
			return fmt.Errorf("Error getting AWS Identity Store Group: %s", err)
		}
		d.SetId(groupID)
		d.Set("display_name", resp.DisplayName)
	} else if displayName != "" {
		log.Printf("[DEBUG] Reading AWS Identity Store Group")
		resp, err := conn.ListGroups(&identitystore.ListGroupsInput{
			IdentityStoreId: aws.String(identityStoreID),
			Filters: []*identitystore.Filter{
				&identitystore.Filter{
					AttributePath:  aws.String("DisplayName"),
					AttributeValue: aws.String(displayName),
				},
			},
		})
		if err != nil {
			return fmt.Errorf("Error getting AWS Identity Store Group: %s", err)
		}
		if resp == nil || len(resp.Groups) == 0 {
			return fmt.Errorf("No AWS Identity Store Group found")
		}
		group := resp.Groups[0]
		d.SetId(aws.StringValue(group.GroupId))
		d.Set("group_id", group.GroupId)
	} else {
		return fmt.Errorf("One of group_id or display_name is required")
	}

	return nil
}
