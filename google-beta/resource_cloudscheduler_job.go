package google

import (
	"encoding/base64"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"google.golang.org/api/cloudscheduler/v1beta1"
)

/*

https://cloud.google.com/scheduler/docs/reference/rest/v1beta1/projects.locations.jobs

TODO
- field validation
- import/update/create/read/delete/exists
- tests
- documentation (website)

*/

type cloudSchedulerJobId struct {
	Project string
	Region  string
	Name    string
}

func (s *cloudSchedulerJobId) jobId() string {
	return fmt.Sprintf("projects/%s/locations/%s/jobs/%s", s.Project, s.Region, s.Name)
}

func (s *cloudSchedulerJobId) locationId() string {
	return fmt.Sprintf("projects/%s/locations/%s", s.Project, s.Region)
}

func (s *cloudSchedulerJobId) terraformId() string {
	return fmt.Sprintf("%s/%s/%s", s.Project, s.Region, s.Name)
}

func parseJobId(id string, config *Config) (*cloudSchedulerJobId, error) {
	parts := strings.Split(id, "/")

	cloudSchedulerJobIdRegex := regexp.MustCompile("^([a-z0-9-]+)/([a-z0-9-])+/([a-zA-Z0-9_-]{1,63})$")

	if cloudSchedulerJobIdRegex.MatchString(id) {
		return &cloudSchedulerJobId{
			Project: parts[0],
			Region:  parts[1],
			Name:    parts[2],
		}, nil
	}

	return nil, fmt.Errorf("Invalid Cloud Scheduler Job id format, expecting " +
		"`{projectId}/{regionId}/{jobName}`")
}

func resourceCloudSchedulerJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudSchedulerJobCreate,
		Read:   resourceCloudSchedulerJobRead,
		Delete: resourceCloudSchedulerJobDelete,
		Exists: resourceCloudSchedulerJobExists,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 500),
			},

			"schedule": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"time_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"retry_config": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
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
				ForceNew:      true,
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
				ForceNew:      true,
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
							ForceNew: true,
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
				ForceNew:      true,
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

			"project": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func expandPubsubTarget(configured []interface{}) *cloudscheduler.PubsubTarget {
	if len(configured) == 0 || configured[0] == nil {
		return &cloudscheduler.PubsubTarget{}
	}

	data := configured[0].(map[string]interface{})

	attributes := make(map[string]string)
	for k, val := range data["attributes"].(map[string]interface{}) {
		attributes[k] = val.(string)
	}

	return &cloudscheduler.PubsubTarget{
		Attributes: attributes,
		Data:       data["data"].(string),
		TopicName:  data["topic_name"].(string),
	}
}

func flattenPubsubTarget(pubsubTarget *cloudscheduler.PubsubTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)
	if pubsubTarget == nil {
		return nil
	}

	result = append(result, map[string]interface{}{
		"attributes": pubsubTarget.Attributes,
		"data":       pubsubTarget.Data,
		"topic_name": pubsubTarget.TopicName,
	})

	return result
}

func expandAppEngineRouting(configured []interface{}) *cloudscheduler.AppEngineRouting {
	if len(configured) == 0 || configured[0] == nil {
		return &cloudscheduler.AppEngineRouting{}
	}

	data := configured[0].(map[string]interface{})
	return &cloudscheduler.AppEngineRouting{
		Instance: data["instance"].(string),
		Service:  data["service"].(string),
		Version:  data["version"].(string),
	}
}

func flattenAppEngineRouting(appEngineRouting *cloudscheduler.AppEngineRouting) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)
	if appEngineRouting == nil {
		return nil
	}

	result = append(result, map[string]interface{}{
		"instance": appEngineRouting.Instance,
		"service":  appEngineRouting.Service,
		"version":  appEngineRouting.Version,
	})

	return result
}

func expandAppEngineHttpTarget(configured []interface{}) *cloudscheduler.AppEngineHttpTarget {
	if len(configured) == 0 || configured[0] == nil {
		return &cloudscheduler.AppEngineHttpTarget{}
	}

	data := configured[0].(map[string]interface{})

	headers := make(map[string]string)
	for k, val := range data["headers"].(map[string]interface{}) {
		headers[k] = val.(string)
	}
	return &cloudscheduler.AppEngineHttpTarget{
		Body:             base64.StdEncoding.EncodeToString([]byte(data["body"].(string))),
		Headers:          headers,
		HttpMethod:       data["http_method"].(string),
		RelativeUri:      data["relative_uri"].(string),
		AppEngineRouting: expandAppEngineRouting(data["app_engine_routing"].([]interface{})),
	}
}

func flattenAppEngineHttpTarget(appEngineHttpTarget *cloudscheduler.AppEngineHttpTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)
	if appEngineHttpTarget == nil {
		return nil
	}

	result = append(result, map[string]interface{}{
		"headers":            appEngineHttpTarget.Headers,
		"http_method":        appEngineHttpTarget.HttpMethod,
		"relative_uri":       appEngineHttpTarget.RelativeUri,
		"app_engine_routing": flattenAppEngineRouting(appEngineHttpTarget.AppEngineRouting),
	})

	return result
}

func expandHttpTarget(configured []interface{}) *cloudscheduler.HttpTarget {
	if len(configured) == 0 || configured[0] == nil {
		return &cloudscheduler.HttpTarget{}
	}

	data := configured[0].(map[string]interface{})

	headers := make(map[string]string)
	for k, val := range data["headers"].(map[string]interface{}) {
		headers[k] = val.(string)
	}
	return &cloudscheduler.HttpTarget{
		Body:       base64.StdEncoding.EncodeToString([]byte(data["body"].(string))),
		Headers:    headers,
		HttpMethod: data["http_method"].(string),
		Uri:        data["uri"].(string),
	}
}

func flattenHttpTarget(httpTarget *cloudscheduler.HttpTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)
	if httpTarget == nil {
		return nil
	}

	result = append(result, map[string]interface{}{
		"uri":         httpTarget.Uri,
		"headers":     httpTarget.Headers,
		"http_method": httpTarget.HttpMethod,
		"body":        httpTarget.Body,
	})

	return result
}

func expandRetryConfig(configured []interface{}) *cloudscheduler.RetryConfig {
	if len(configured) == 0 || configured[0] == nil {
		return &cloudscheduler.RetryConfig{}
	}

	data := configured[0].(map[string]interface{})
	return &cloudscheduler.RetryConfig{
		MaxBackoffDuration: data["max_backoff_duration"].(string),
		MaxDoublings:       int64(data["max_doublings"].(int)),
		MaxRetryDuration:   data["max_retry_duration"].(string),
		MinBackoffDuration: data["min_backoff_duration"].(string),
		RetryCount:         int64(data["retry_count"].(int)),
	}
}

func flattenRetryConfig(retryConfig *cloudscheduler.RetryConfig) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)
	if retryConfig == nil {
		return nil
	}

	result = append(result, map[string]interface{}{
		"retry_count":          retryConfig.RetryCount,
		"max_retry_duration":   retryConfig.MaxRetryDuration,
		"min_backoff_duration": retryConfig.MinBackoffDuration,
		"max_backoff_duration": retryConfig.MaxBackoffDuration,
		"max_doublings":        retryConfig.MaxDoublings,
	})

	return result
}

func resourceCloudSchedulerJobCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	region, err := getRegion(d, config)
	if err != nil {
		return err
	}

	jobId := &cloudSchedulerJobId{
		Project: project,
		Region:  region,
		Name:    d.Get("name").(string),
	}

	job := &cloudscheduler.Job{
		Name:            jobId.jobId(),
		Schedule:        d.Get("schedule").(string),
		ForceSendFields: []string{},
	}

	if v, ok := d.GetOk("description"); ok {
		job.Description = v.(string)
	}

	if v, ok := d.GetOk("pubsub_target"); ok {
		job.PubsubTarget = expandPubsubTarget(v.([]interface{}))
	} else if v, ok := d.GetOk("app_engine_http_target"); ok {
		job.AppEngineHttpTarget = expandAppEngineHttpTarget(v.([]interface{}))
	} else if v, ok := d.GetOk("http_target"); ok {
		job.HttpTarget = expandHttpTarget(v.([]interface{}))
	} else {
		return fmt.Errorf("One of `pubsub_target`, `app_engine_http_target` or `http_target` is required: " +
			"You must specify a target when deploying a new scheduler job.")
	}

	if v, ok := d.GetOk("time_zone"); ok {
		job.TimeZone = v.(string)
	}

	if v, ok := d.GetOk("retry_config"); ok {
		job.RetryConfig = expandRetryConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating scheduler job: %s", job.Name)
	op, err := config.clientCloudScheduler.Projects.Locations.Jobs.Create(
		jobId.locationId(), job).Do()
	_ = op
	if err != nil {
		return err
	}

	// Name of job should be unique
	d.SetId(jobId.terraformId())

	return resourceCloudSchedulerJobRead(d, meta)
}

func resourceCloudSchedulerJobRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	jobId, err := parseJobId(d.Id(), config)
	if err != nil {
		return err
	}

	job, err := config.clientCloudScheduler.Projects.Locations.Jobs.Get(jobId.jobId()).Do()
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Target CloudScheduler Job %q", jobId.Name))
	}

	d.Set("name", jobId.Name)
	d.Set("description", job.Description)
	d.Set("schedule", job.Schedule)

	if job.PubsubTarget != nil {
		d.Set("pubsub_target", flattenPubsubTarget(job.PubsubTarget))
	} else if job.HttpTarget != nil {
		d.Set("pubsub_target", flattenHttpTarget(job.HttpTarget))
	} else if job.AppEngineHttpTarget != nil {
		d.Set("pubsub_target", flattenAppEngineHttpTarget(job.AppEngineHttpTarget))
	}

	if job.RetryConfig != nil {
		d.Set("retry_config", flattenRetryConfig(job.RetryConfig))
	}

	if job.PubsubTarget != nil {
		d.Set("time_zone", job.TimeZone)
	}
	d.Set("region", jobId.Region)
	d.Set("project", jobId.Project)

	return nil
}

func resourceCloudSchedulerJobDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	jobId, err := parseJobId(d.Id(), config)
	if err != nil {
		return err
	}

	op, err := config.clientCloudScheduler.Projects.Locations.Jobs.Delete(jobId.jobId()).Do()
	_ = op
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceCloudSchedulerJobExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	config := meta.(*Config)

	jobId, err := parseJobId(d.Id(), config)
	if err != nil {
		return false, err
	}

	_, err = config.clientCloudScheduler.Projects.Locations.Jobs.Get(jobId.jobId()).Do()
	if err != nil {
		if err = handleNotFoundError(err, d, fmt.Sprintf("CloudScheduler Job %s", jobId.Name)); err == nil {
			return false, nil
		}
		// There was some other error in reading the resource
		return true, err
	}
	return true, nil
}
