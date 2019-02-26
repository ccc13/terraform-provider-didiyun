package didiyun

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	didiyun "github.com/shonenada/didiyun-go"
	ds "github.com/shonenada/didiyun-go/schema"
	sshkey "github.com/shonenada/didiyun-go/sshkey"
)

func resourceDidiyunSSHKey() *schema.Resource {
	return &schema.Resource{
		Read:   resourceDidiyunSSHKeyRead,
		Create: resourceDidiyunSSHKeyCreate,
		Update: resourceDidiyunSSHKeyUpdate,
		Delete: resourceDidiyunSSHKeyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "name of this SSH Key",
				ValidateFunc: validation.NoZeroValues,
			},
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// Didiyun does not supported get sshkey by id,
// so we need list all keys, then find the key by id.
func resourceDidiyunSSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	keys, err := client.SSHKey().List()
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Failed to request SSH Keys: %v", err)
	}

	for _, ele := range keys {
		if ele.PubkeyUuid == id {
			d.Set("name", ele.Name)
			d.Set("key", ele.Key)
			d.Set("fingerprint", ele.Fingerprint)
			return nil
		}
	}

	return fmt.Errorf("Failed to find SSH Keys: %v", err)
	d.SetId("")
	return nil
}

func resourceDidiyunSSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).Client()

	req := sshkey.CreateRequest{
		Name: d.Get("name").(string),
		Key:  d.Get("key").(string),
	}

	response, err := client.SSHKey().Create(&req)
	if err != nil {
		return fmt.Errorf("Failed to create SSH Key: %v", err)
	}

	data := response[0]

	d.SetId(data.PubkeyUuid)

	return resourceDidiyunSSHKeyRead(d, meta)
}

func resourceDidiyunSSHKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDidiyunSSHKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).Client()

	req := sshkey.DeleteRequest{
		PubkeyUuid: d.Id(),
	}

	_, err := client.SSHKey().Delete(&req)

	if err != nil {
		return fmt.Errorf("Failed to delete SSH Key: %v", err)
	}

	d.SetId("")
	return nil
}
