package google

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

/*

https://cloud.google.com/scheduler/docs/reference/rest/v1beta1/projects.locations.jobs

TODO
- validation
- import
- update
- create
- read
- delete
- exists

*/

func resourceCloudSchedulerJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudSchedulerJobCreate,
		Read:   resourceCloudSchedulerJobRead,
		Delete: resourceCloudSchedulerJobDelete,
		Exists: resourceCloudSchedulerJobExists,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 500),
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"retry_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retry_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
						"max_retry_duration": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"min_backoff_duration": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"max_backoff_duration": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"max_doublings": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"pubsub_target": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"app_engine_http_target", "http_target"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"topic_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"data": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"attributes": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
			"app_engine_http_target": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"pubsub_target", "http_target"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_method": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"app_engine_routing": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"service": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"version": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"instance": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"relative_uri": {
							Type:     schema.TypeString,
							Required: true,
						},
						"headers": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"body": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"http_target": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"pubsub_target", "app_engine_http_target"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:     schema.TypeString,
							Required: true,
						},
						"http_method": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"headers": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"body": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceCloudSchedulerJobCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCloudSchedulerJobRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCloudSchedulerJobDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCloudSchedulerJobExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return false, nil
}
