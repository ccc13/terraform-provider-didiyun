package didiyun

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	didyun "github.com/shonenada/didiyun-go"
	dc2 "github.com/shonenada/didiyun-go/dc2"
	ds "github.com/shonenada/didiyun-go/schema"
)

func flattenDidiyunTags(tags []string) *schema.Set {
	flattentags := schema.NewSet(schema.HashString, []interface{}{})
	for _, v := range tags {
		flattentags.add(v)
	}
	return flattentags
}

func flattenDidiyunEip(eip *dc.Eip) map[string]interface{} {
	result := map[string]string{
		"ip_address": eip.Ip,
	}
	return result
}

func flattenDidiyunEbs(ebs *dc.Ebs) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)
	for _, v := range ebs {
		r := make(map[string]interface{})
		r["attr"] = v.Attr
		r["name"] = v.Name
		r["size"] = v.Size
		r["disktype"] = v.Spec.DiskType
		r["tags"] = v.EbsTags

		result = append(result, r)
	}
	return result
}

func resourceDidiyunDC2() *schema.Resource {
	return &schema.Resource{
		Read:   resourceDidiyunDC2Read,
		Create: resourceDidiyunDC2Create,
		Update: resourceDidiyunDC2Update,
		Delete: resourceDidiyunDC2Delete,

		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     false,
				ValidateFunc: validation.NoZeroValues,
			},
			"auto_continue": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"pay_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"subnet_uuid": {
				Type:     schema.String,
				Optional: true,
			},
			"dc2_model": {
				Type:        schema.String,
				Optional:    true,
				Description: "Model of DC2, all avaliable models: https://docs.didiyun.com/static/docs-content/products/DC2/%E5%88%9B%E5%BB%BADC2%EF%BC%88CreateDC2%EF%BC%89.html#Dc2Models",
			},
			"image_uuid": {
				Type:     schema.String,
				ForceNew: true,
				Optional: true,
			},
			"snap_uuid": {
				Type:     schema.String,
				ForceNew: true,
				Optional: true,
			},
			"sshkkeys": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
				Description: "List of uuids of SSHKey",
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rootdisk_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rootdisk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(40, 500),
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
			},
			"sg_uuids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
			},
			"ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Intranet IP",
			},
			"os_type": {
				Type:     schema.String,
				Computed: true,
			},
			"eip": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"band_width": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"charge_with_flow": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"tags": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"ip_address": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
			"ebs": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					"count": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  1,
					},
					"attr": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"size": {
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(20, 16384),
					},
					"disktype": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"snap_uuid": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"tags": {
						Type: schema.TypeSet,
						Elem: &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceDidiyunDC2Read(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).Client()

	uuid := d.Id()
	regionId := d.Get("region_id").(string)

	req := dc2.GetRequest{
		RegionId: regionId,
		Dc2Uuid:  uuid,
	}

	r, err := client.Dc2().Get(&req)
	if err != nil {
		return fmt.Errorf("Failed to read Dc2: %v", err)
	}

	data := r[0]

	d.Set("name", data.Name)
	d.Set("ip", data.Ip)
	d.Set("os_type", data.OSType)
	d.Set("region_id", data.Region.Id)
	d.Set("zone_id", data.Region.Zone.Id)
	d.Set("eip", flattenDidiyunEip(data.Eip))
	d.Set("ebs", flattenDidiyunEbs(data.Ebs))

	if err := d.Set("tags", flattenDidiyunTags(data.Tags)); err != nil {
		return fmt.Errorf("Failed to set `tags`: %v", err)
	}

	return nil
}

func resourceDidiyunDC2Create(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).Client()
	req := dc2.CreateRequest{
		RegionId:     d.Get("region_id"),
		ZoneId:       d.Get("zone_id"),
		Name:         d.Get("name"),
		AutoContinue: d.Get("auto_continue"),
		PayPeriod:    d.Get("pay_period"),
		Count:        d.Get("count"),
		SubnetUuid:   d.Get("subnet_uuid"),
		Dc2Model:     d.Get("dc2_model"),
		ImgUuid:      d.Get("image_uuid"),
		PubKeyUuids:  d.Get("sshkeys"),
		Password:     d.Get("password"),
		RootDiskType: d.Get("rootdisk_type"),
		RootDiskSize: d.Get("rootdisk_size"),
		Dc2Tags:      d.Get("tags"),
		SgUuids:      d.Get("sg_uuids"),
		Eip:          d.Get("eip"),
		Ebs:          d.Get("ebs"),
	}

	response, err := client.Dc2().Create(*req)

	if err != nil {
		return fmt.Errorf("Failed to create DC2: %v", err)
	}

	data := response[0]

	d.SetId(data.ResourceUuid)

	return resourceDidiyunDC2Read(d, meta)
}

func resourceDidiyunDC2Update(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDidiyunDC2Delete(d *schema.ResourceData, meta interface{}) error {
	return nil
}