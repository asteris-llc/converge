package fastly

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	gofastly "github.com/sethvargo/go-fastly"
)

var fastlyNoServiceFoundErr = errors.New("No matching Fastly Service found")

func resourceServiceV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceV1Create,
		Read:   resourceServiceV1Read,
		Update: resourceServiceV1Update,
		Delete: resourceServiceV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name for this Service",
			},

			// Active Version represents the currently activated version in Fastly. In
			// Terraform, we abstract this number away from the users and manage
			// creating and activating. It's used internally, but also exported for
			// users to see.
			"active_version": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"domain": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The domain that this Service will respond to",
						},

						"comment": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"condition": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"statement": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The statement used to determine if the condition is met",
							StateFunc: func(v interface{}) string {
								value := v.(string)
								// Trim newlines and spaces, to match Fastly API
								return strings.TrimSpace(value)
							},
						},
						"priority": &schema.Schema{
							Type:        schema.TypeInt,
							Required:    true,
							Description: "A number used to determine the order in which multiple conditions execute. Lower numbers execute first",
						},
						"type": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Type of the condition, either `REQUEST`, `RESPONSE`, or `CACHE`",
						},
					},
				},
			},

			"default_ttl": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3600,
				Description: "The default Time-to-live (TTL) for the version",
			},

			"default_host": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The default hostname for the version",
			},

			"backend": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// required fields
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name for this Backend",
						},
						"address": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "An IPv4, hostname, or IPv6 address for the Backend",
						},
						// Optional fields, defaults where they exist
						"auto_loadbalance": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Should this Backend be load balanced",
						},
						"between_bytes_timeout": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     10000,
							Description: "How long to wait between bytes in milliseconds",
						},
						"connect_timeout": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1000,
							Description: "How long to wait for a timeout in milliseconds",
						},
						"error_threshold": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Number of errors to allow before the Backend is marked as down",
						},
						"first_byte_timeout": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     15000,
							Description: "How long to wait for the first bytes in milliseconds",
						},
						"max_conn": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     200,
							Description: "Maximum number of connections for this Backend",
						},
						"port": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     80,
							Description: "The port number Backend responds on. Default 80",
						},
						"ssl_check_cert": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Be strict on checking SSL certs",
						},
						// UseSSL is something we want to support in the future, but
						// requires SSL setup we don't yet have
						// TODO: Provide all SSL fields from https://docs.fastly.com/api/config#backend
						// "use_ssl": &schema.Schema{
						// 	Type:        schema.TypeBool,
						// 	Optional:    true,
						// 	Default:     false,
						// 	Description: "Whether or not to use SSL to reach the Backend",
						// },
						"weight": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     100,
							Description: "The portion of traffic to send to a specific origins. Each origin receives weight/total of the traffic.",
						},
					},
				},
			},

			"force_destroy": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},

			"cache_setting": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// required fields
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name to refer to this Cache Setting",
						},
						"cache_condition": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Condition to check if this Cache Setting applies",
						},
						"action": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Action to take",
						},
						// optional
						"stale_ttl": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Max 'Time To Live' for stale (unreachable) objects.",
							Default:     300,
						},
						"ttl": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The 'Time To Live' for the object",
						},
					},
				},
			},

			"gzip": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// required fields
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name to refer to this gzip condition",
						},
						// optional fields
						"content_types": &schema.Schema{
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Content types to apply automatic gzip to",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"extensions": &schema.Schema{
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "File extensions to apply automatic gzip to. Do not include '.'",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						// These fields represent Fastly options that Terraform does not
						// currently support
						"cache_condition": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Optional name of a CacheCondition to apply.",
						},
					},
				},
			},

			"header": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// required fields
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name to refer to this Header object",
						},
						"action": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "One of set, append, delete, regex, or regex_repeat",
							ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
								var found bool
								for _, t := range []string{"set", "append", "delete", "regex", "regex_repeat"} {
									if v.(string) == t {
										found = true
									}
								}
								if !found {
									es = append(es, fmt.Errorf(
										"Fastly Header action is case sensitive and must be one of 'set', 'append', 'delete', 'regex', or 'regex_repeat'; found: %s", v.(string)))
								}
								return
							},
						},
						"type": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Type to manipulate: request, fetch, cache, response",
							ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
								var found bool
								for _, t := range []string{"request", "fetch", "cache", "response"} {
									if v.(string) == t {
										found = true
									}
								}
								if !found {
									es = append(es, fmt.Errorf(
										"Fastly Header type is case sensitive and must be one of 'request', 'fetch', 'cache', or 'response'; found: %s", v.(string)))
								}
								return
							},
						},
						"destination": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Header this affects",
						},
						// Optional fields, defaults where they exist
						"ignore_if_set": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Don't add the header if it is already. (Only applies to 'set' action.). Default `false`",
						},
						"source": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Variable to be used as a source for the header content (Does not apply to 'delete' action.)",
						},
						"regex": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Regular expression to use (Only applies to 'regex' and 'regex_repeat' actions.)",
						},
						"substitution": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Value to substitute in place of regular expression. (Only applies to 'regex' and 'regex_repeat'.)",
						},
						"priority": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     100,
							Description: "Lower priorities execute first. (Default: 100.)",
						},
						// These fields represent Fastly options that Terraform does not
						// currently support
						"request_condition": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Optional name of a RequestCondition to apply.",
						},
						"cache_condition": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Optional name of a CacheCondition to apply.",
						},
						"response_condition": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Optional name of a ResponseCondition to apply.",
						},
					},
				},
			},

			"s3logging": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this logging setup",
						},
						"bucket_name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "S3 Bucket name to store logs in",
						},
						"s3_access_key": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_S3_ACCESS_KEY", ""),
							Description: "AWS Access Key",
						},
						"s3_secret_key": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("FASTLY_S3_SECRET_KEY", ""),
							Description: "AWS Secret Key",
						},
						// Optional fields
						"path": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Path to store the files. Must end with a trailing slash",
						},
						"domain": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Bucket endpoint",
						},
						"gzip_level": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Gzip Compression level",
						},
						"period": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     3600,
							Description: "How frequently the logs should be transferred, in seconds (Default 3600)",
						},
						"format": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%h %l %u %t %r %>s",
							Description: "Apache-style string or VCL variables to use for log formatting",
						},
						"timestamp_format": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "%Y-%m-%dT%H:%M:%S.000",
							Description: "specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)",
						},
					},
				},
			},

			"request_setting": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Required fields
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name to refer to this Request Setting",
						},
						"request_condition": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of a RequestCondition to apply.",
						},
						// Optional fields
						"max_stale_age": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     60,
							Description: "How old an object is allowed to be, in seconds. Default `60`",
						},
						"force_miss": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Force a cache miss for the request",
						},
						"force_ssl": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Forces the request use SSL",
						},
						"action": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Allows you to terminate request handling and immediately perform an action",
						},
						"bypass_busy_wait": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Disable collapsed forwarding",
						},
						"hash_keys": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Comma separated list of varnish request object fields that should be in the hash key",
						},
						"xff": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "append",
							Description: "X-Forwarded-For options",
						},
						"timer_support": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Injects the X-Timer info into the request",
						},
						"geo_headers": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Inject Fastly-Geo-Country, Fastly-Geo-City, and Fastly-Geo-Region",
						},
						"default_host": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "the host header",
						},
					},
				},
			},
			"vcl": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "A name to refer to this VCL configuration",
						},
						"content": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The contents of this VCL configuration",
							StateFunc: func(v interface{}) string {
								switch v.(type) {
								case string:
									hash := sha1.Sum([]byte(v.(string)))
									return hex.EncodeToString(hash[:])
								default:
									return ""
								}
							},
						},
						"main": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Should this VCL configuation be the main configuration",
						},
					},
				},
			},
		},
	}
}

func resourceServiceV1Create(d *schema.ResourceData, meta interface{}) error {
	if err := validateVCLs(d); err != nil {
		return err
	}

	conn := meta.(*FastlyClient).conn
	service, err := conn.CreateService(&gofastly.CreateServiceInput{
		Name:    d.Get("name").(string),
		Comment: "Managed by Terraform",
	})

	if err != nil {
		return err
	}

	d.SetId(service.ID)
	return resourceServiceV1Update(d, meta)
}

func resourceServiceV1Update(d *schema.ResourceData, meta interface{}) error {
	if err := validateVCLs(d); err != nil {
		return err
	}

	conn := meta.(*FastlyClient).conn

	// Update Name. No new verions is required for this
	if d.HasChange("name") {
		_, err := conn.UpdateService(&gofastly.UpdateServiceInput{
			ID:   d.Id(),
			Name: d.Get("name").(string),
		})
		if err != nil {
			return err
		}
	}

	// Once activated, Versions are locked and become immutable. This is true for
	// versions that are no longer active. For Domains, Backends, DefaultHost and
	// DefaultTTL, a new Version must be created first, and updates posted to that
	// Version. Loop these attributes and determine if we need to create a new version first
	var needsChange bool
	for _, v := range []string{
		"domain",
		"backend",
		"default_host",
		"default_ttl",
		"header",
		"gzip",
		"s3logging",
		"condition",
		"request_setting",
		"cache_setting",
		"vcl",
	} {
		if d.HasChange(v) {
			needsChange = true
		}
	}

	if needsChange {
		latestVersion := d.Get("active_version").(string)
		if latestVersion == "" {
			// If the service was just created, there is an empty Version 1 available
			// that is unlocked and can be updated
			latestVersion = "1"
		} else {
			// Clone the latest version, giving us an unlocked version we can modify
			log.Printf("[DEBUG] Creating clone of version (%s) for updates", latestVersion)
			newVersion, err := conn.CloneVersion(&gofastly.CloneVersionInput{
				Service: d.Id(),
				Version: latestVersion,
			})
			if err != nil {
				return err
			}

			// The new version number is named "Number", but it's actually a string
			latestVersion = newVersion.Number

			// New versions are not immediately found in the API, or are not
			// immediately mutable, so we need to sleep a few and let Fastly ready
			// itself. Typically, 7 seconds is enough
			log.Printf("[DEBUG] Sleeping 7 seconds to allow Fastly Version to be available")
			time.Sleep(7 * time.Second)
		}

		// update general settings
		if d.HasChange("default_host") || d.HasChange("default_ttl") {
			opts := gofastly.UpdateSettingsInput{
				Service: d.Id(),
				Version: latestVersion,
				// default_ttl has the same default value of 3600 that is provided by
				// the Fastly API, so it's safe to include here
				DefaultTTL: uint(d.Get("default_ttl").(int)),
			}

			if attr, ok := d.GetOk("default_host"); ok {
				opts.DefaultHost = attr.(string)
			}

			log.Printf("[DEBUG] Update Settings opts: %#v", opts)
			_, err := conn.UpdateSettings(&opts)
			if err != nil {
				return err
			}
		}

		// Conditions need to be updated first, as they can be referenced by other
		// configuraiton objects (Backends, Request Headers, etc)

		// Find difference in Conditions
		if d.HasChange("condition") {
			// Note: we don't utilize the PUT endpoint to update these objects, we simply
			// destroy any that have changed, and create new ones with the updated
			// values. This is how Terraform works with nested sub resources, we only
			// get the full diff not a partial set item diff. Because this is done
			// on a new version of the Fastly Service configuration, this is considered safe

			oc, nc := d.GetChange("condition")
			if oc == nil {
				oc = new(schema.Set)
			}
			if nc == nil {
				nc = new(schema.Set)
			}

			ocs := oc.(*schema.Set)
			ncs := nc.(*schema.Set)
			removeConditions := ocs.Difference(ncs).List()
			addConditions := ncs.Difference(ocs).List()

			// DELETE old Conditions
			for _, cRaw := range removeConditions {
				cf := cRaw.(map[string]interface{})
				opts := gofastly.DeleteConditionInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    cf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Conditions Removal opts: %#v", opts)
				err := conn.DeleteCondition(&opts)
				if err != nil {
					return err
				}
			}

			// POST new Conditions
			for _, cRaw := range addConditions {
				cf := cRaw.(map[string]interface{})
				opts := gofastly.CreateConditionInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    cf["name"].(string),
					Type:    cf["type"].(string),
					// need to trim leading/tailing spaces, incase the config has HEREDOC
					// formatting and contains a trailing new line
					Statement: strings.TrimSpace(cf["statement"].(string)),
					Priority:  cf["priority"].(int),
				}

				log.Printf("[DEBUG] Create Conditions Opts: %#v", opts)
				_, err := conn.CreateCondition(&opts)
				if err != nil {
					return err
				}
			}
		}

		// Find differences in domains
		if d.HasChange("domain") {
			od, nd := d.GetChange("domain")
			if od == nil {
				od = new(schema.Set)
			}
			if nd == nil {
				nd = new(schema.Set)
			}

			ods := od.(*schema.Set)
			nds := nd.(*schema.Set)

			remove := ods.Difference(nds).List()
			add := nds.Difference(ods).List()

			// Delete removed domains
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteDomainInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Domain removal opts: %#v", opts)
				err := conn.DeleteDomain(&opts)
				if err != nil {
					return err
				}
			}

			// POST new Domains
			for _, dRaw := range add {
				df := dRaw.(map[string]interface{})
				opts := gofastly.CreateDomainInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				if v, ok := df["comment"]; ok {
					opts.Comment = v.(string)
				}

				log.Printf("[DEBUG] Fastly Domain Addition opts: %#v", opts)
				_, err := conn.CreateDomain(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in backends
		if d.HasChange("backend") {
			ob, nb := d.GetChange("backend")
			if ob == nil {
				ob = new(schema.Set)
			}
			if nb == nil {
				nb = new(schema.Set)
			}

			obs := ob.(*schema.Set)
			nbs := nb.(*schema.Set)
			removeBackends := obs.Difference(nbs).List()
			addBackends := nbs.Difference(obs).List()

			// DELETE old Backends
			for _, bRaw := range removeBackends {
				bf := bRaw.(map[string]interface{})
				opts := gofastly.DeleteBackendInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    bf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Backend removal opts: %#v", opts)
				err := conn.DeleteBackend(&opts)
				if err != nil {
					return err
				}
			}

			// Find and post new Backends
			for _, dRaw := range addBackends {
				df := dRaw.(map[string]interface{})
				opts := gofastly.CreateBackendInput{
					Service:             d.Id(),
					Version:             latestVersion,
					Name:                df["name"].(string),
					Address:             df["address"].(string),
					AutoLoadbalance:     df["auto_loadbalance"].(bool),
					SSLCheckCert:        df["ssl_check_cert"].(bool),
					Port:                uint(df["port"].(int)),
					BetweenBytesTimeout: uint(df["between_bytes_timeout"].(int)),
					ConnectTimeout:      uint(df["connect_timeout"].(int)),
					ErrorThreshold:      uint(df["error_threshold"].(int)),
					FirstByteTimeout:    uint(df["first_byte_timeout"].(int)),
					MaxConn:             uint(df["max_conn"].(int)),
					Weight:              uint(df["weight"].(int)),
				}

				log.Printf("[DEBUG] Create Backend Opts: %#v", opts)
				_, err := conn.CreateBackend(&opts)
				if err != nil {
					return err
				}
			}
		}

		if d.HasChange("header") {
			oh, nh := d.GetChange("header")
			if oh == nil {
				oh = new(schema.Set)
			}
			if nh == nil {
				nh = new(schema.Set)
			}

			ohs := oh.(*schema.Set)
			nhs := nh.(*schema.Set)

			remove := ohs.Difference(nhs).List()
			add := nhs.Difference(ohs).List()

			// Delete removed headers
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteHeaderInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Header removal opts: %#v", opts)
				err := conn.DeleteHeader(&opts)
				if err != nil {
					return err
				}
			}

			// POST new Headers
			for _, dRaw := range add {
				opts, err := buildHeader(dRaw.(map[string]interface{}))
				if err != nil {
					log.Printf("[DEBUG] Error building Header: %s", err)
					return err
				}
				opts.Service = d.Id()
				opts.Version = latestVersion

				log.Printf("[DEBUG] Fastly Header Addition opts: %#v", opts)
				_, err = conn.CreateHeader(opts)
				if err != nil {
					return err
				}
			}
		}

		// Find differences in Gzips
		if d.HasChange("gzip") {
			og, ng := d.GetChange("gzip")
			if og == nil {
				og = new(schema.Set)
			}
			if ng == nil {
				ng = new(schema.Set)
			}

			ogs := og.(*schema.Set)
			ngs := ng.(*schema.Set)

			remove := ogs.Difference(ngs).List()
			add := ngs.Difference(ogs).List()

			// Delete removed gzip rules
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteGzipInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Gzip removal opts: %#v", opts)
				err := conn.DeleteGzip(&opts)
				if err != nil {
					return err
				}
			}

			// POST new Gzips
			for _, dRaw := range add {
				df := dRaw.(map[string]interface{})
				opts := gofastly.CreateGzipInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				if v, ok := df["content_types"]; ok {
					if len(v.(*schema.Set).List()) > 0 {
						var cl []string
						for _, c := range v.(*schema.Set).List() {
							cl = append(cl, c.(string))
						}
						opts.ContentTypes = strings.Join(cl, " ")
					}
				}

				if v, ok := df["extensions"]; ok {
					if len(v.(*schema.Set).List()) > 0 {
						var el []string
						for _, e := range v.(*schema.Set).List() {
							el = append(el, e.(string))
						}
						opts.Extensions = strings.Join(el, " ")
					}
				}

				log.Printf("[DEBUG] Fastly Gzip Addition opts: %#v", opts)
				_, err := conn.CreateGzip(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in s3logging
		if d.HasChange("s3logging") {
			os, ns := d.GetChange("s3logging")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			oss := os.(*schema.Set)
			nss := ns.(*schema.Set)
			removeS3Logging := oss.Difference(nss).List()
			addS3Logging := nss.Difference(oss).List()

			// DELETE old S3 Log configurations
			for _, sRaw := range removeS3Logging {
				sf := sRaw.(map[string]interface{})
				opts := gofastly.DeleteS3Input{
					Service: d.Id(),
					Version: latestVersion,
					Name:    sf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly S3 Logging removal opts: %#v", opts)
				err := conn.DeleteS3(&opts)
				if err != nil {
					return err
				}
			}

			// POST new/updated S3 Logging
			for _, sRaw := range addS3Logging {
				sf := sRaw.(map[string]interface{})

				// Fastly API will not error if these are omitted, so we throw an error
				// if any of these are empty
				for _, sk := range []string{"s3_access_key", "s3_secret_key"} {
					if sf[sk].(string) == "" {
						return fmt.Errorf("[ERR] No %s found for S3 Log stream setup for Service (%s)", sk, d.Id())
					}
				}

				opts := gofastly.CreateS3Input{
					Service:         d.Id(),
					Version:         latestVersion,
					Name:            sf["name"].(string),
					BucketName:      sf["bucket_name"].(string),
					AccessKey:       sf["s3_access_key"].(string),
					SecretKey:       sf["s3_secret_key"].(string),
					Period:          uint(sf["period"].(int)),
					GzipLevel:       uint(sf["gzip_level"].(int)),
					Domain:          sf["domain"].(string),
					Path:            sf["path"].(string),
					Format:          sf["format"].(string),
					TimestampFormat: sf["timestamp_format"].(string),
				}

				log.Printf("[DEBUG] Create S3 Logging Opts: %#v", opts)
				_, err := conn.CreateS3(&opts)
				if err != nil {
					return err
				}
			}
		}

		// find difference in request settings
		if d.HasChange("request_setting") {
			os, ns := d.GetChange("request_setting")
			if os == nil {
				os = new(schema.Set)
			}
			if ns == nil {
				ns = new(schema.Set)
			}

			ors := os.(*schema.Set)
			nrs := ns.(*schema.Set)
			removeRequestSettings := ors.Difference(nrs).List()
			addRequestSettings := nrs.Difference(ors).List()

			// DELETE old Request Settings configurations
			for _, sRaw := range removeRequestSettings {
				sf := sRaw.(map[string]interface{})
				opts := gofastly.DeleteRequestSettingInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    sf["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Request Setting removal opts: %#v", opts)
				err := conn.DeleteRequestSetting(&opts)
				if err != nil {
					return err
				}
			}

			// POST new/updated Request Setting
			for _, sRaw := range addRequestSettings {
				opts, err := buildRequestSetting(sRaw.(map[string]interface{}))
				if err != nil {
					log.Printf("[DEBUG] Error building Requset Setting: %s", err)
					return err
				}
				opts.Service = d.Id()
				opts.Version = latestVersion

				log.Printf("[DEBUG] Create Request Setting Opts: %#v", opts)
				_, err = conn.CreateRequestSetting(opts)
				if err != nil {
					return err
				}
			}
		}

		// Find differences in VCLs
		if d.HasChange("vcl") {
			// Note: as above with Gzip and S3 logging, we don't utilize the PUT
			// endpoint to update a VCL, we simply destroy it and create a new one.
			oldVCLVal, newVCLVal := d.GetChange("vcl")
			if oldVCLVal == nil {
				oldVCLVal = new(schema.Set)
			}
			if newVCLVal == nil {
				newVCLVal = new(schema.Set)
			}

			oldVCLSet := oldVCLVal.(*schema.Set)
			newVCLSet := newVCLVal.(*schema.Set)

			remove := oldVCLSet.Difference(newVCLSet).List()
			add := newVCLSet.Difference(oldVCLSet).List()

			// Delete removed VCL configurations
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteVCLInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly VCL Removal opts: %#v", opts)
				err := conn.DeleteVCL(&opts)
				if err != nil {
					return err
				}
			}
			// POST new VCL configurations
			for _, dRaw := range add {
				df := dRaw.(map[string]interface{})
				opts := gofastly.CreateVCLInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
					Content: df["content"].(string),
				}

				log.Printf("[DEBUG] Fastly VCL Addition opts: %#v", opts)
				_, err := conn.CreateVCL(&opts)
				if err != nil {
					return err
				}

				// if this new VCL is the main
				if df["main"].(bool) {
					opts := gofastly.ActivateVCLInput{
						Service: d.Id(),
						Version: latestVersion,
						Name:    df["name"].(string),
					}
					log.Printf("[DEBUG] Fastly VCL activation opts: %#v", opts)
					_, err := conn.ActivateVCL(&opts)
					if err != nil {
						return err
					}

				}
			}
		}

		// Find differences in Cache Settings
		if d.HasChange("cache_setting") {
			oc, nc := d.GetChange("cache_setting")
			if oc == nil {
				oc = new(schema.Set)
			}
			if nc == nil {
				nc = new(schema.Set)
			}

			ocs := oc.(*schema.Set)
			ncs := nc.(*schema.Set)

			remove := ocs.Difference(ncs).List()
			add := ncs.Difference(ocs).List()

			// Delete removed Cache Settings
			for _, dRaw := range remove {
				df := dRaw.(map[string]interface{})
				opts := gofastly.DeleteCacheSettingInput{
					Service: d.Id(),
					Version: latestVersion,
					Name:    df["name"].(string),
				}

				log.Printf("[DEBUG] Fastly Cache Settings removal opts: %#v", opts)
				err := conn.DeleteCacheSetting(&opts)
				if err != nil {
					return err
				}
			}

			// POST new Cache Settings
			for _, dRaw := range add {
				opts, err := buildCacheSetting(dRaw.(map[string]interface{}))
				if err != nil {
					log.Printf("[DEBUG] Error building Cache Setting: %s", err)
					return err
				}
				opts.Service = d.Id()
				opts.Version = latestVersion

				log.Printf("[DEBUG] Fastly Cache Settings Addition opts: %#v", opts)
				_, err = conn.CreateCacheSetting(opts)
				if err != nil {
					return err
				}
			}
		}

		// validate version
		log.Printf("[DEBUG] Validating Fastly Service (%s), Version (%s)", d.Id(), latestVersion)
		valid, msg, err := conn.ValidateVersion(&gofastly.ValidateVersionInput{
			Service: d.Id(),
			Version: latestVersion,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error checking validation: %s", err)
		}

		if !valid {
			return fmt.Errorf("[ERR] Invalid configuration for Fastly Service (%s): %s", d.Id(), msg)
		}

		log.Printf("[DEBUG] Activating Fastly Service (%s), Version (%s)", d.Id(), latestVersion)
		_, err = conn.ActivateVersion(&gofastly.ActivateVersionInput{
			Service: d.Id(),
			Version: latestVersion,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error activating version (%s): %s", latestVersion, err)
		}

		// Only if the version is valid and activated do we set the active_version.
		// This prevents us from getting stuck in cloning an invalid version
		d.Set("active_version", latestVersion)
	}

	return resourceServiceV1Read(d, meta)
}

func resourceServiceV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	// Find the Service. Discard the service because we need the ServiceDetails,
	// not just a Service record
	_, err := findService(d.Id(), meta)
	if err != nil {
		switch err {
		case fastlyNoServiceFoundErr:
			log.Printf("[WARN] %s for ID (%s)", err, d.Id())
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	s, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	d.Set("name", s.Name)
	d.Set("active_version", s.ActiveVersion.Number)

	// If CreateService succeeds, but initial updates to the Service fail, we'll
	// have an empty ActiveService version (no version is active, so we can't
	// query for information on it)
	if s.ActiveVersion.Number != "" {
		settingsOpts := gofastly.GetSettingsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		}
		if settings, err := conn.GetSettings(&settingsOpts); err == nil {
			d.Set("default_host", settings.DefaultHost)
			d.Set("default_ttl", settings.DefaultTTL)
		} else {
			return fmt.Errorf("[ERR] Error looking up Version settings for (%s), version (%s): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		// TODO: update go-fastly to support an ActiveVersion struct, which contains
		// domain and backend info in the response. Here we do 2 additional queries
		// to find out that info
		log.Printf("[DEBUG] Refreshing Domains for (%s)", d.Id())
		domainList, err := conn.ListDomains(&gofastly.ListDomainsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Domains for (%s), version (%s): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		// Refresh Domains
		dl := flattenDomains(domainList)

		if err := d.Set("domain", dl); err != nil {
			log.Printf("[WARN] Error setting Domains for (%s): %s", d.Id(), err)
		}

		// Refresh Backends
		log.Printf("[DEBUG] Refreshing Backends for (%s)", d.Id())
		backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Backends for (%s), version (%s): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		bl := flattenBackends(backendList)

		if err := d.Set("backend", bl); err != nil {
			log.Printf("[WARN] Error setting Backends for (%s): %s", d.Id(), err)
		}

		// refresh headers
		log.Printf("[DEBUG] Refreshing Headers for (%s)", d.Id())
		headerList, err := conn.ListHeaders(&gofastly.ListHeadersInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Headers for (%s), version (%s): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		hl := flattenHeaders(headerList)

		if err := d.Set("header", hl); err != nil {
			log.Printf("[WARN] Error setting Headers for (%s): %s", d.Id(), err)
		}

		// refresh gzips
		log.Printf("[DEBUG] Refreshing Gzips for (%s)", d.Id())
		gzipsList, err := conn.ListGzips(&gofastly.ListGzipsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Gzips for (%s), version (%s): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		gl := flattenGzips(gzipsList)

		if err := d.Set("gzip", gl); err != nil {
			log.Printf("[WARN] Error setting Gzips for (%s): %s", d.Id(), err)
		}

		// refresh S3 Logging
		log.Printf("[DEBUG] Refreshing S3 Logging for (%s)", d.Id())
		s3List, err := conn.ListS3s(&gofastly.ListS3sInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up S3 Logging for (%s), version (%s): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		sl := flattenS3s(s3List)

		if err := d.Set("s3logging", sl); err != nil {
			log.Printf("[WARN] Error setting S3 Logging for (%s): %s", d.Id(), err)
		}

		// refresh Conditions
		log.Printf("[DEBUG] Refreshing Conditions for (%s)", d.Id())
		conditionList, err := conn.ListConditions(&gofastly.ListConditionsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Conditions for (%s), version (%s): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		cl := flattenConditions(conditionList)

		if err := d.Set("condition", cl); err != nil {
			log.Printf("[WARN] Error setting Conditions for (%s): %s", d.Id(), err)
		}

		// refresh Request Settings
		log.Printf("[DEBUG] Refreshing Request Settings for (%s)", d.Id())
		rsList, err := conn.ListRequestSettings(&gofastly.ListRequestSettingsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Request Settings for (%s), version (%s): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		rl := flattenRequestSettings(rsList)

		if err := d.Set("request_setting", rl); err != nil {
			log.Printf("[WARN] Error setting Request Settings for (%s): %s", d.Id(), err)
		}

		// refresh VCLs
		log.Printf("[DEBUG] Refreshing VCLs for (%s)", d.Id())
		vclList, err := conn.ListVCLs(&gofastly.ListVCLsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up VCLs for (%s), version (%s): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		vl := flattenVCLs(vclList)

		if err := d.Set("vcl", vl); err != nil {
			log.Printf("[WARN] Error setting VCLs for (%s): %s", d.Id(), err)
		}

		// refresh Cache Settings
		log.Printf("[DEBUG] Refreshing Cache Settings for (%s)", d.Id())
		cslList, err := conn.ListCacheSettings(&gofastly.ListCacheSettingsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Cache Settings for (%s), version (%s): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		csl := flattenCacheSettings(cslList)

		if err := d.Set("cache_setting", csl); err != nil {
			log.Printf("[WARN] Error setting Cache Settings for (%s): %s", d.Id(), err)
		}

	} else {
		log.Printf("[DEBUG] Active Version for Service (%s) is empty, no state to refresh", d.Id())
	}

	return nil
}

func resourceServiceV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	// Fastly will fail to delete any service with an Active Version.
	// If `force_destroy` is given, we deactivate the active version and then send
	// the DELETE call
	if d.Get("force_destroy").(bool) {
		s, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
			ID: d.Id(),
		})

		if err != nil {
			return err
		}

		if s.ActiveVersion.Number != "" {
			_, err := conn.DeactivateVersion(&gofastly.DeactivateVersionInput{
				Service: d.Id(),
				Version: s.ActiveVersion.Number,
			})
			if err != nil {
				return err
			}
		}
	}

	err := conn.DeleteService(&gofastly.DeleteServiceInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	_, err = findService(d.Id(), meta)
	if err != nil {
		switch err {
		// we expect no records to be found here
		case fastlyNoServiceFoundErr:
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	// findService above returned something and nil error, but shouldn't have
	return fmt.Errorf("[WARN] Tried deleting Service (%s), but was still found", d.Id())

}

func flattenDomains(list []*gofastly.Domain) []map[string]interface{} {
	dl := make([]map[string]interface{}, 0, len(list))

	for _, d := range list {
		dl = append(dl, map[string]interface{}{
			"name":    d.Name,
			"comment": d.Comment,
		})
	}

	return dl
}

func flattenBackends(backendList []*gofastly.Backend) []map[string]interface{} {
	var bl []map[string]interface{}
	for _, b := range backendList {
		// Convert Backend to a map for saving to state.
		nb := map[string]interface{}{
			"name":                  b.Name,
			"address":               b.Address,
			"auto_loadbalance":      b.AutoLoadbalance,
			"between_bytes_timeout": int(b.BetweenBytesTimeout),
			"connect_timeout":       int(b.ConnectTimeout),
			"error_threshold":       int(b.ErrorThreshold),
			"first_byte_timeout":    int(b.FirstByteTimeout),
			"max_conn":              int(b.MaxConn),
			"port":                  int(b.Port),
			"ssl_check_cert":        b.SSLCheckCert,
			"weight":                int(b.Weight),
		}

		bl = append(bl, nb)
	}
	return bl
}

// findService finds a Fastly Service via the ListServices endpoint, returning
// the Service if found.
//
// Fastly API does not include any "deleted_at" type parameter to indicate
// that a Service has been deleted. GET requests to a deleted Service will
// return 200 OK and have the full output of the Service for an unknown time
// (days, in my testing). In order to determine if a Service is deleted, we
// need to hit /service and loop the returned Services, searching for the one
// in question. This endpoint only returns active or "alive" services. If the
// Service is not included, then it's "gone"
//
// Returns a fastlyNoServiceFoundErr error if the Service is not found in the
// ListServices response.
func findService(id string, meta interface{}) (*gofastly.Service, error) {
	conn := meta.(*FastlyClient).conn

	l, err := conn.ListServices(&gofastly.ListServicesInput{})
	if err != nil {
		return nil, fmt.Errorf("[WARN] Error listing services when deleting Fastly Service (%s): %s", id, err)
	}

	for _, s := range l {
		if s.ID == id {
			log.Printf("[DEBUG] Found Service (%s)", id)
			return s, nil
		}
	}

	return nil, fastlyNoServiceFoundErr
}

func flattenHeaders(headerList []*gofastly.Header) []map[string]interface{} {
	var hl []map[string]interface{}
	for _, h := range headerList {
		// Convert Header to a map for saving to state.
		nh := map[string]interface{}{
			"name":               h.Name,
			"action":             h.Action,
			"ignore_if_set":      h.IgnoreIfSet,
			"type":               h.Type,
			"destination":        h.Destination,
			"source":             h.Source,
			"regex":              h.Regex,
			"substitution":       h.Substitution,
			"priority":           int(h.Priority),
			"request_condition":  h.RequestCondition,
			"cache_condition":    h.CacheCondition,
			"response_condition": h.ResponseCondition,
		}

		for k, v := range nh {
			if v == "" {
				delete(nh, k)
			}
		}

		hl = append(hl, nh)
	}
	return hl
}

func buildHeader(headerMap interface{}) (*gofastly.CreateHeaderInput, error) {
	df := headerMap.(map[string]interface{})
	opts := gofastly.CreateHeaderInput{
		Name:              df["name"].(string),
		IgnoreIfSet:       gofastly.Compatibool(df["ignore_if_set"].(bool)),
		Destination:       df["destination"].(string),
		Priority:          uint(df["priority"].(int)),
		Source:            df["source"].(string),
		Regex:             df["regex"].(string),
		Substitution:      df["substitution"].(string),
		RequestCondition:  df["request_condition"].(string),
		CacheCondition:    df["cache_condition"].(string),
		ResponseCondition: df["response_condition"].(string),
	}

	act := strings.ToLower(df["action"].(string))
	switch act {
	case "set":
		opts.Action = gofastly.HeaderActionSet
	case "append":
		opts.Action = gofastly.HeaderActionAppend
	case "delete":
		opts.Action = gofastly.HeaderActionDelete
	case "regex":
		opts.Action = gofastly.HeaderActionRegex
	case "regex_repeat":
		opts.Action = gofastly.HeaderActionRegexRepeat
	}

	ty := strings.ToLower(df["type"].(string))
	switch ty {
	case "request":
		opts.Type = gofastly.HeaderTypeRequest
	case "fetch":
		opts.Type = gofastly.HeaderTypeFetch
	case "cache":
		opts.Type = gofastly.HeaderTypeCache
	case "response":
		opts.Type = gofastly.HeaderTypeResponse
	}

	return &opts, nil
}

func buildCacheSetting(cacheMap interface{}) (*gofastly.CreateCacheSettingInput, error) {
	df := cacheMap.(map[string]interface{})
	opts := gofastly.CreateCacheSettingInput{
		Name:           df["name"].(string),
		StaleTTL:       uint(df["stale_ttl"].(int)),
		CacheCondition: df["cache_condition"].(string),
	}

	if v, ok := df["ttl"]; ok {
		opts.TTL = uint(v.(int))
	}

	act := strings.ToLower(df["action"].(string))
	switch act {
	case "cache":
		opts.Action = gofastly.CacheSettingActionCache
	case "pass":
		opts.Action = gofastly.CacheSettingActionPass
	case "restart":
		opts.Action = gofastly.CacheSettingActionRestart
	}

	return &opts, nil
}

func flattenGzips(gzipsList []*gofastly.Gzip) []map[string]interface{} {
	var gl []map[string]interface{}
	for _, g := range gzipsList {
		// Convert Gzip to a map for saving to state.
		ng := map[string]interface{}{
			"name":            g.Name,
			"cache_condition": g.CacheCondition,
		}

		if g.Extensions != "" {
			e := strings.Split(g.Extensions, " ")
			var et []interface{}
			for _, ev := range e {
				et = append(et, ev)
			}
			ng["extensions"] = schema.NewSet(schema.HashString, et)
		}

		if g.ContentTypes != "" {
			c := strings.Split(g.ContentTypes, " ")
			var ct []interface{}
			for _, cv := range c {
				ct = append(ct, cv)
			}
			ng["content_types"] = schema.NewSet(schema.HashString, ct)
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ng {
			if v == "" {
				delete(ng, k)
			}
		}

		gl = append(gl, ng)
	}

	return gl
}

func flattenS3s(s3List []*gofastly.S3) []map[string]interface{} {
	var sl []map[string]interface{}
	for _, s := range s3List {
		// Convert S3s to a map for saving to state.
		ns := map[string]interface{}{
			"name":             s.Name,
			"bucket_name":      s.BucketName,
			"s3_access_key":    s.AccessKey,
			"s3_secret_key":    s.SecretKey,
			"path":             s.Path,
			"period":           s.Period,
			"domain":           s.Domain,
			"gzip_level":       s.GzipLevel,
			"format":           s.Format,
			"timestamp_format": s.TimestampFormat,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ns {
			if v == "" {
				delete(ns, k)
			}
		}

		sl = append(sl, ns)
	}

	return sl
}

func flattenConditions(conditionList []*gofastly.Condition) []map[string]interface{} {
	var cl []map[string]interface{}
	for _, c := range conditionList {
		// Convert Conditions to a map for saving to state.
		nc := map[string]interface{}{
			"name":      c.Name,
			"statement": c.Statement,
			"type":      c.Type,
			"priority":  c.Priority,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nc {
			if v == "" {
				delete(nc, k)
			}
		}

		cl = append(cl, nc)
	}

	return cl
}

func flattenRequestSettings(rsList []*gofastly.RequestSetting) []map[string]interface{} {
	var rl []map[string]interface{}
	for _, r := range rsList {
		// Convert Request Settings to a map for saving to state.
		nrs := map[string]interface{}{
			"name":              r.Name,
			"max_stale_age":     r.MaxStaleAge,
			"force_miss":        r.ForceMiss,
			"force_ssl":         r.ForceSSL,
			"action":            r.Action,
			"bypass_busy_wait":  r.BypassBusyWait,
			"hash_keys":         r.HashKeys,
			"xff":               r.XForwardedFor,
			"timer_support":     r.TimerSupport,
			"geo_headers":       r.GeoHeaders,
			"default_host":      r.DefaultHost,
			"request_condition": r.RequestCondition,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nrs {
			if v == "" {
				delete(nrs, k)
			}
		}

		rl = append(rl, nrs)
	}

	return rl
}

func buildRequestSetting(requestSettingMap interface{}) (*gofastly.CreateRequestSettingInput, error) {
	df := requestSettingMap.(map[string]interface{})
	opts := gofastly.CreateRequestSettingInput{
		Name:             df["name"].(string),
		MaxStaleAge:      uint(df["max_stale_age"].(int)),
		ForceMiss:        gofastly.Compatibool(df["force_miss"].(bool)),
		ForceSSL:         gofastly.Compatibool(df["force_ssl"].(bool)),
		BypassBusyWait:   gofastly.Compatibool(df["bypass_busy_wait"].(bool)),
		HashKeys:         df["hash_keys"].(string),
		TimerSupport:     gofastly.Compatibool(df["timer_support"].(bool)),
		GeoHeaders:       gofastly.Compatibool(df["geo_headers"].(bool)),
		DefaultHost:      df["default_host"].(string),
		RequestCondition: df["request_condition"].(string),
	}

	act := strings.ToLower(df["action"].(string))
	switch act {
	case "lookup":
		opts.Action = gofastly.RequestSettingActionLookup
	case "pass":
		opts.Action = gofastly.RequestSettingActionPass
	}

	xff := strings.ToLower(df["xff"].(string))
	switch xff {
	case "clear":
		opts.XForwardedFor = gofastly.RequestSettingXFFClear
	case "leave":
		opts.XForwardedFor = gofastly.RequestSettingXFFLeave
	case "append":
		opts.XForwardedFor = gofastly.RequestSettingXFFAppend
	case "append_all":
		opts.XForwardedFor = gofastly.RequestSettingXFFAppendAll
	case "overwrite":
		opts.XForwardedFor = gofastly.RequestSettingXFFOverwrite
	}

	return &opts, nil
}

func flattenCacheSettings(csList []*gofastly.CacheSetting) []map[string]interface{} {
	var csl []map[string]interface{}
	for _, cl := range csList {
		// Convert Cache Settings to a map for saving to state.
		clMap := map[string]interface{}{
			"name":            cl.Name,
			"action":          cl.Action,
			"cache_condition": cl.CacheCondition,
			"stale_ttl":       cl.StaleTTL,
			"ttl":             cl.TTL,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range clMap {
			if v == "" {
				delete(clMap, k)
			}
		}

		csl = append(csl, clMap)
	}

	return csl
}

func flattenVCLs(vclList []*gofastly.VCL) []map[string]interface{} {
	var vl []map[string]interface{}
	for _, vcl := range vclList {
		// Convert VCLs to a map for saving to state.
		vclMap := map[string]interface{}{
			"name":    vcl.Name,
			"content": vcl.Content,
			"main":    vcl.Main,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range vclMap {
			if v == "" {
				delete(vclMap, k)
			}
		}

		vl = append(vl, vclMap)
	}

	return vl
}

func validateVCLs(d *schema.ResourceData) error {
	// TODO: this would be nice to move into a resource/collection validation function, once that is available
	// (see https://github.com/hashicorp/terraform/pull/4348 and https://github.com/hashicorp/terraform/pull/6508)
	vcls, exists := d.GetOk("vcl")
	if !exists {
		return nil
	}

	numberOfMainVCLs, numberOfIncludeVCLs := 0, 0
	for _, vclElem := range vcls.(*schema.Set).List() {
		vcl := vclElem.(map[string]interface{})
		if mainVal, hasMain := vcl["main"]; hasMain && mainVal.(bool) {
			numberOfMainVCLs++
		} else {
			numberOfIncludeVCLs++
		}
	}
	if numberOfMainVCLs == 0 && numberOfIncludeVCLs > 0 {
		return fmt.Errorf("if you include VCL configurations, one of them should have main = true")
	}
	if numberOfMainVCLs > 1 {
		return fmt.Errorf("you cannot have more than one VCL configuration with main = true")
	}
	return nil
}
