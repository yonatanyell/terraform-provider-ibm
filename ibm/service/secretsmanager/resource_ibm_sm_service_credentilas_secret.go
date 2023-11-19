// Copyright IBM Corp. 2023 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package secretsmanager

import (
	"context"
	"fmt"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/secrets-manager-go-sdk/v2/secretsmanagerv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
	"strings"
)

func ResourceIbmSmServiceCredentialsSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIbmSmServiceCredentialsSecretCreate,
		ReadContext:   resourceIbmSmServiceCredentialsSecretRead,
		UpdateContext: resourceIbmSmServiceCredentialsSecretUpdate,
		DeleteContext: resourceIbmSmServiceCredentialsSecretDelete,
		Importer:      &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"secret_type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The secret type. Supported types are arbitrary, certificates (imported, public, and private), IAM credentials, key-value, and user credentials.",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "A human-readable name to assign to your secret.To protect your privacy, do not use personal data, such as your name or location, as a name for your secret.",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An extended description of your secret.To protect your privacy, do not use personal data, such as your name or location, as a description for your secret group.",
			},
			"secret_group_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "A v4 UUID identifier, or `default` secret group.",
			},
			"labels": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "Labels that you can use to search for secrets in your instance.Up to 30 labels can be created.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"custom_metadata": &schema.Schema{
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true,
				Description: "The secret metadata that a user can customize.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"version_custom_metadata": &schema.Schema{
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true,
				Description: "The secret version metadata that a user can customize.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"created_by": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier that is associated with the entity that created the secret.",
			},
			"created_at": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date when a resource was created. The date format follows RFC 3339.",
			},
			"credentials": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The properties of the service credentials secret payload.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apikey": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "The API key that is generated for this secret.",
						},
						"cos_hmac_keys": &schema.Schema{
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The Cloud Object Storage HMAC keys that are returned after you create a service credentials secret.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_key_id": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The access key ID for Cloud Object Storage HMAC credentials.",
									},
									"secret_access_key": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The secret access key ID for Cloud Object Storage HMAC credentials.",
									},
								},
							},
						},
						"endpoints": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The endpoints that are returned after you create a service credentials secret.",
						},
						"iam_apikey_description": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the generated IAM API key.",
						},
						"iam_apikey_name": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the generated IAM API key.",
						},
						"iam_role_crn": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IAM role CRN that is returned after you create a service credentials secret.",
						},
						"iam_serviceid_crn": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IAM serviceId CRN that is returned after you create a service credentials secret.",
						},
						"resource_instance_id": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource instance CRN that is returned after you create a service credentials secret.",
						},
					},
				},
			},
			"crn": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A CRN that uniquely identifies an IBM Cloud resource.",
			},
			"downloaded": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether the secret data that is associated with a secret version was retrieved in a call to the service API.",
			},
			"locks_total": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of locks of the secret.",
			},
			"next_rotation_date": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date that the secret is scheduled for automatic rotation. The service automatically creates a new version of the secret on its next rotation date. This field exists only for secrets that have an existing rotation policy.",
			},
			"rotation": &schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Determines whether Secrets Manager rotates your secrets automatically.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_rotate": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Determines whether Secrets Manager rotates your secret automatically.Default is `false`. If `auto_rotate` is set to `true` the service rotates your secret based on the defined interval.",
						},
						"interval": &schema.Schema{
							Type:             schema.TypeInt,
							Optional:         true,
							Computed:         true,
							Description:      "The length of the secret rotation time interval.",
							DiffSuppressFunc: rotationAttributesDiffSuppress,
						},
						"unit": &schema.Schema{
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							Description:      "The units for the secret rotation time interval.",
							DiffSuppressFunc: rotationAttributesDiffSuppress,
						},
					},
				},
			},
			"source_service": &schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				ForceNew:    true,
				Description: "The properties required for creating the service credentials for the specified source service instance.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance": &schema.Schema{
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "The source service instance identifier.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"crn": &schema.Schema{
										Type:        schema.TypeString,
										Required:    true,
										Description: "A CRN that uniquely identifies a service credentials target.",
									},
								},
							},
						},
						"role": &schema.Schema{
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "The service-specific custom role object, CRN role is accepted. Refer to the service’s documentation for supported roles.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"crn": &schema.Schema{
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
										Description: "The CRN role identifier for creating a service-id.",
									},
								},
							},
						},
						"iam": &schema.Schema{
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The source service IAM data is returned in case IAM credentials where created for this secret.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"apikey": &schema.Schema{
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The IAM apikey metadata for the IAM credentials that were generated.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": &schema.Schema{
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The IAM API key name for the generated service credentials.",
												},
												"description": &schema.Schema{
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The IAM API key description for the generated service credentials.",
												},
											},
										},
									},
									"role": &schema.Schema{
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The IAM role for the generate service credentials.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"crn": &schema.Schema{
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The IAM role CRN assigned to the generated service credentials.",
												},
											},
										},
									},
									"serviceid": &schema.Schema{
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The IAM serviceid for the generated service credentials.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"crn": &schema.Schema{
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The IAM Service ID CRN.",
												},
											},
										},
									},
								},
							},
						},
						"resource_key": &schema.Schema{
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The source service resource key data of the generated service credentials.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"crn": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The resource key CRN of the generated service credentials.",
									},
									"name": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The resource key name of the generated service credentials.",
									},
								},
							},
						},
						"parameters": &schema.Schema{
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "The collection of parameters for the service credentials target.",
						},
					},
				},
			},
			"state": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The secret state that is based on NIST SP 800-57. States are integers and correspond to the `Pre-activation = 0`, `Active = 1`,  `Suspended = 2`, `Deactivated = 3`, and `Destroyed = 5` values.",
			},
			"state_description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A text representation of the secret state.",
			},
			"ttl": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: StringIsIntBetween(86400, 7776000),
				Description:  "The time-to-live (TTL) or lease duration to assign to generated credentials.",
			},
			"updated_at": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date when a resource was recently modified. The date format follows RFC 3339.",
			},
			"versions_total": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of versions of the secret.",
			},
			"secret_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A v4 UUID identifier.",
			},
		},
	}
}

func resourceIbmSmServiceCredentialsSecretCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretsManagerClient, err := meta.(conns.ClientSession).SecretsManagerV2()
	if err != nil {
		return diag.FromErr(err)
	}

	region := getRegion(secretsManagerClient, d)
	instanceId := d.Get("instance_id").(string)
	secretsManagerClient = getClientWithInstanceEndpoint(secretsManagerClient, instanceId, region, getEndpointType(secretsManagerClient, d))

	createSecretOptions := &secretsmanagerv2.CreateSecretOptions{}

	secretPrototypeModel, err := resourceIbmSmServiceCredentialsSecretMapToSecretPrototype(d)
	if err != nil {
		return diag.FromErr(err)
	}
	createSecretOptions.SetSecretPrototype(secretPrototypeModel)

	secretIntf, response, err := secretsManagerClient.CreateSecretWithContext(context, createSecretOptions)
	if err != nil {
		log.Printf("[DEBUG] CreateSecretWithContext failed %s\n%s", err, response)
		return diag.FromErr(fmt.Errorf("CreateSecretWithContext failed %s\n%s", err, response))
	}

	secret := secretIntf.(*secretsmanagerv2.ServiceCredentialsSecret)
	d.SetId(fmt.Sprintf("%s/%s/%s", region, instanceId, *secret.ID))
	d.Set("secret_id", *secret.ID)

	return resourceIbmSmServiceCredentialsSecretRead(context, d, meta)
}

func resourceIbmSmServiceCredentialsSecretRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretsManagerClient, err := meta.(conns.ClientSession).SecretsManagerV2()
	if err != nil {
		return diag.FromErr(err)
	}

	id := strings.Split(d.Id(), "/")
	if len(id) != 3 {
		return diag.Errorf("Wrong format of resource ID. To import a secret use the format `<region>/<instance_id>/<secret_id>`")
	}
	region := id[0]
	instanceId := id[1]
	secretId := id[2]
	secretsManagerClient = getClientWithInstanceEndpoint(secretsManagerClient, instanceId, region, getEndpointType(secretsManagerClient, d))

	getSecretOptions := &secretsmanagerv2.GetSecretOptions{}

	getSecretOptions.SetID(secretId)

	secretIntf, response, err := secretsManagerClient.GetSecretWithContext(context, getSecretOptions)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] GetSecretWithContext failed %s\n%s", err, response)
		return diag.FromErr(fmt.Errorf("GetSecretWithContext failed %s\n%s", err, response))
	}

	secret := secretIntf.(*secretsmanagerv2.ServiceCredentialsSecret)

	if err = d.Set("secret_id", secretId); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting secret_id: %s", err))
	}
	if err = d.Set("instance_id", instanceId); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting instance_id: %s", err))
	}
	if err = d.Set("region", region); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting region: %s", err))
	}
	if err = d.Set("created_by", secret.CreatedBy); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting created_by: %s", err))
	}
	if err = d.Set("created_at", DateTimeToRFC3339(secret.CreatedAt)); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting created_at: %s", err))
	}
	if err = d.Set("crn", secret.Crn); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting crn: %s", err))
	}
	if err = d.Set("downloaded", secret.Downloaded); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting downloaded: %s", err))
	}
	if err = d.Set("locks_total", flex.IntValue(secret.LocksTotal)); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting locks_total: %s", err))
	}
	if err = d.Set("name", secret.Name); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting name: %s", err))
	}
	if err = d.Set("secret_group_id", secret.SecretGroupID); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting secret_group_id: %s", err))
	}
	if err = d.Set("secret_type", secret.SecretType); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting secret_type: %s", err))
	}
	if err = d.Set("state", flex.IntValue(secret.State)); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting state: %s", err))
	}
	if err = d.Set("state_description", secret.StateDescription); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting state_description: %s", err))
	}
	if err = d.Set("updated_at", DateTimeToRFC3339(secret.UpdatedAt)); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting updated_at: %s", err))
	}
	if err = d.Set("versions_total", flex.IntValue(secret.VersionsTotal)); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting versions_total: %s", err))
	}
	if secret.CustomMetadata != nil {
		d.Set("custom_metadata", secret.CustomMetadata)
	}
	if err = d.Set("description", secret.Description); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting description: %s", err))
	}
	if secret.Labels != nil {
		if err = d.Set("labels", secret.Labels); err != nil {
			return diag.FromErr(fmt.Errorf("Error setting labels: %s", err))
		}
	}
	rotationMap, err := resourceIbmSmServiceCredentialsSecretRotationPolicyToMap(secret.Rotation)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(rotationMap) > 0 {
		if err = d.Set("rotation", []map[string]interface{}{rotationMap}); err != nil {
			return diag.FromErr(fmt.Errorf("Error setting rotation: %s", err))
		}
	}
	sourceServiceMap, err := resourceIbmSmServiceCredentialsSecretSourceServiceToMap(secret.SourceService)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(sourceServiceMap) > 0 {
		if err = d.Set("source_service", []map[string]interface{}{sourceServiceMap}); err != nil {
			return diag.FromErr(fmt.Errorf("Error setting source_service: %s", err))
		}
	}
	if secret.Credentials != nil {
		credentialsMap, err := resourceIbmSmServiceCredentialsSecretCredentialsToMap(secret.Credentials)
		if err != nil {
			return diag.FromErr(err)
		}
		if len(credentialsMap) > 0 {
			if err = d.Set("credentials", []map[string]interface{}{credentialsMap}); err != nil {
				return diag.FromErr(fmt.Errorf("Error setting credentialsMap: %s", err))
			}
		}
	}
	if err = d.Set("next_rotation_date", DateTimeToRFC3339(secret.NextRotationDate)); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting next_rotation_date: %s", err))
	}

	// Call get version metadata API to get the current version_custom_metadata
	getVersionMetdataOptions := &secretsmanagerv2.GetSecretVersionMetadataOptions{}
	getVersionMetdataOptions.SetSecretID(secretId)
	getVersionMetdataOptions.SetID("current")

	versionMetadataIntf, response, err := secretsManagerClient.GetSecretVersionMetadataWithContext(context, getVersionMetdataOptions)
	if err != nil {
		log.Printf("[DEBUG] GetSecretVersionMetadataWithContext failed %s\n%s", err, response)
		return diag.FromErr(fmt.Errorf("GetSecretVersionMetadataWithContext failed %s\n%s", err, response))
	}

	versionMetadata := versionMetadataIntf.(*secretsmanagerv2.ServiceCredentialsSecretVersionMetadata)
	if versionMetadata.VersionCustomMetadata != nil {
		if err = d.Set("version_custom_metadata", versionMetadata.VersionCustomMetadata); err != nil {
			return diag.FromErr(fmt.Errorf("Error setting version_custom_metadata: %s", err))
		}
	}

	return nil
}

func resourceIbmSmServiceCredentialsSecretUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretsManagerClient, err := meta.(conns.ClientSession).SecretsManagerV2()
	if err != nil {
		return diag.FromErr(err)
	}

	id := strings.Split(d.Id(), "/")
	region := id[0]
	instanceId := id[1]
	secretId := id[2]
	secretsManagerClient = getClientWithInstanceEndpoint(secretsManagerClient, instanceId, region, getEndpointType(secretsManagerClient, d))

	updateSecretMetadataOptions := &secretsmanagerv2.UpdateSecretMetadataOptions{}

	updateSecretMetadataOptions.SetID(secretId)

	hasChange := false

	patchVals := &secretsmanagerv2.SecretMetadataPatch{}

	if d.HasChange("name") {
		patchVals.Name = core.StringPtr(d.Get("name").(string))
		hasChange = true
	}
	if d.HasChange("description") {
		patchVals.Description = core.StringPtr(d.Get("description").(string))
		hasChange = true
	}
	if d.HasChange("ttl") {
		patchVals.TTL = core.StringPtr(d.Get("ttl").(string))
		hasChange = true
	}
	if d.HasChange("labels") {
		labels := d.Get("labels").([]interface{})
		labelsParsed := make([]string, len(labels))
		for i, v := range labels {
			labelsParsed[i] = fmt.Sprint(v)
		}
		patchVals.Labels = labelsParsed
		hasChange = true
	}
	if d.HasChange("custom_metadata") {
		patchVals.CustomMetadata = d.Get("custom_metadata").(map[string]interface{})
		hasChange = true
	}
	if d.HasChange("rotation") {
		RotationModel, err := resourceIbmSmServiceCredentialsSecretMapToRotationPolicy(d.Get("rotation").([]interface{})[0].(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] UpdateSecretMetadataWithContext failed: Reading Rotation parameter failed: %s", err)
			return diag.FromErr(fmt.Errorf("UpdateSecretMetadataWithContext failed: Reading Rotation parameter failed: %s", err))
		}
		patchVals.Rotation = RotationModel
		hasChange = true
	}

	// Apply change in metadata (if changed)
	if hasChange {
		updateSecretMetadataOptions.SecretMetadataPatch, _ = patchVals.AsPatch()
		_, response, err := secretsManagerClient.UpdateSecretMetadataWithContext(context, updateSecretMetadataOptions)
		if err != nil {
			log.Printf("[DEBUG] UpdateSecretMetadataWithContext failed %s\n%s", err, response)
			return diag.FromErr(fmt.Errorf("UpdateSecretMetadataWithContext failed %s\n%s", err, response))
		}
	}

	if d.HasChange("version_custom_metadata") {
		// Apply change to version_custom_metadata in current version
		secretVersionMetadataPatchModel := new(secretsmanagerv2.SecretVersionMetadataPatch)
		secretVersionMetadataPatchModel.VersionCustomMetadata = d.Get("version_custom_metadata").(map[string]interface{})
		secretVersionMetadataPatchModelAsPatch, _ := secretVersionMetadataPatchModel.AsPatch()

		updateSecretVersionOptions := &secretsmanagerv2.UpdateSecretVersionMetadataOptions{}
		updateSecretVersionOptions.SetSecretID(secretId)
		updateSecretVersionOptions.SetID("current")
		updateSecretVersionOptions.SetSecretVersionMetadataPatch(secretVersionMetadataPatchModelAsPatch)
		_, response, err := secretsManagerClient.UpdateSecretVersionMetadataWithContext(context, updateSecretVersionOptions)
		if err != nil {
			if hasChange {
				// Call the read function to update the Terraform state with the change already applied to the metadata
				resourceIbmSmServiceCredentialsSecretRead(context, d, meta)
			}
			log.Printf("[DEBUG] UpdateSecretVersionMetadataWithContext failed %s\n%s", err, response)
			return diag.FromErr(fmt.Errorf("UpdateSecretVersionMetadataWithContext failed %s\n%s", err, response))
		}
	}

	return resourceIbmSmServiceCredentialsSecretRead(context, d, meta)
}

func resourceIbmSmServiceCredentialsSecretDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretsManagerClient, err := meta.(conns.ClientSession).SecretsManagerV2()
	if err != nil {
		return diag.FromErr(err)
	}

	id := strings.Split(d.Id(), "/")
	region := id[0]
	instanceId := id[1]
	secretId := id[2]
	secretsManagerClient = getClientWithInstanceEndpoint(secretsManagerClient, instanceId, region, getEndpointType(secretsManagerClient, d))

	deleteSecretOptions := &secretsmanagerv2.DeleteSecretOptions{}

	deleteSecretOptions.SetID(secretId)

	response, err := secretsManagerClient.DeleteSecretWithContext(context, deleteSecretOptions)
	if err != nil {
		log.Printf("[DEBUG] DeleteSecretWithContext failed %s\n%s", err, response)
		return diag.FromErr(fmt.Errorf("DeleteSecretWithContext failed %s\n%s", err, response))
	}

	d.SetId("")

	return nil
}

func resourceIbmSmServiceCredentialsSecretMapToSecretPrototype(d *schema.ResourceData) (*secretsmanagerv2.ServiceCredentialsSecretPrototype, error) {
	model := &secretsmanagerv2.ServiceCredentialsSecretPrototype{}
	model.SecretType = core.StringPtr("service_credentials")

	if _, ok := d.GetOk("name"); ok {
		model.Name = core.StringPtr(d.Get("name").(string))
	}
	if _, ok := d.GetOk("description"); ok {
		model.Description = core.StringPtr(d.Get("description").(string))
	}
	if _, ok := d.GetOk("secret_group_id"); ok {
		model.SecretGroupID = core.StringPtr(d.Get("secret_group_id").(string))
	}
	if _, ok := d.GetOk("labels"); ok {
		labels := d.Get("labels").([]interface{})
		labelsParsed := make([]string, len(labels))
		for i, v := range labels {
			labelsParsed[i] = fmt.Sprint(v)
		}
		model.Labels = labelsParsed
	}
	if _, ok := d.GetOk("ttl"); ok {
		model.TTL = core.StringPtr(d.Get("ttl").(string))
	}
	if _, ok := d.GetOk("rotation"); ok {
		RotationModel, err := resourceIbmSmServiceCredentialsSecretMapToRotationPolicy(d.Get("rotation").([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return model, err
		}
		model.Rotation = RotationModel
	}
	if _, ok := d.GetOk("source_service"); ok {
		SourceServiceModel, err := resourceIbmSmServiceCredentialsSecretMapToSourceService(d.Get("source_service").([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return model, err
		}
		model.SourceService = SourceServiceModel
	}
	if _, ok := d.GetOk("custom_metadata"); ok {
		model.CustomMetadata = d.Get("custom_metadata").(map[string]interface{})
	}
	if _, ok := d.GetOk("version_custom_metadata"); ok {
		model.VersionCustomMetadata = d.Get("version_custom_metadata").(map[string]interface{})
	}
	return model, nil
}

func resourceIbmSmServiceCredentialsSecretMapToRotationPolicy(modelMap map[string]interface{}) (secretsmanagerv2.RotationPolicyIntf, error) {
	model := &secretsmanagerv2.RotationPolicy{}
	if modelMap["auto_rotate"] != nil {
		model.AutoRotate = core.BoolPtr(modelMap["auto_rotate"].(bool))
	}
	if modelMap["interval"].(int) == 0 {
		model.Interval = nil
	} else {
		model.Interval = core.Int64Ptr(int64(modelMap["interval"].(int)))
	}
	if modelMap["unit"] != nil && modelMap["unit"].(string) != "" {
		model.Unit = core.StringPtr(modelMap["unit"].(string))
	}
	return model, nil
}

func resourceIbmSmServiceCredentialsSecretMapToSourceService(modelMap map[string]interface{}) (*secretsmanagerv2.ServiceCredentialsSecretSourceService, error) {
	mainModel := &secretsmanagerv2.ServiceCredentialsSecretSourceService{}

	if modelMap["instance"] != nil && len(modelMap["instance"].([]interface{})) > 0 {
		instanceModel := &secretsmanagerv2.ServiceCredentialsSourceServiceInstance{}
		if modelMap["instance"].([]interface{})[0].(map[string]interface{})["crn"].(string) != "" {
			instanceModel.Crn = core.StringPtr(modelMap["instance"].([]interface{})[0].(map[string]interface{})["crn"].(string))
			mainModel.Instance = instanceModel
		}
	}

	if modelMap["role"] != nil && len(modelMap["role"].([]interface{})) > 0 {
		roleModel := &secretsmanagerv2.ServiceCredentialsSourceServiceRole{}
		if modelMap["role"].([]interface{})[0].(map[string]interface{})["crn"].(string) != "" {
			roleModel.Crn = core.StringPtr(modelMap["role"].([]interface{})[0].(map[string]interface{})["crn"].(string))
			mainModel.Role = roleModel
		}
	}

	if modelMap["parameters"] != nil {
		mainModel.Parameters = &secretsmanagerv2.ServiceCredentialsSourceServiceParameters{}
		parametersMap := modelMap["parameters"].(map[string]interface{})
		for k, v := range parametersMap {
			if k == "serviceid_crn" {
				serviceIdCrn := v.(string)
				mainModel.Parameters.ServiceidCrn = &serviceIdCrn
			} else if v == "true" || v == "false" {
				b, _ := strconv.ParseBool(v.(string))
				mainModel.Parameters.SetProperty(k, b)
			} else {
				mainModel.Parameters.SetProperty(k, v)
			}
		}
	}
	return mainModel, nil
}

func resourceIbmSmServiceCredentialsSecretRotationPolicyToMap(modelIntf secretsmanagerv2.RotationPolicyIntf) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	model := modelIntf.(*secretsmanagerv2.RotationPolicy)
	if model.AutoRotate != nil {
		modelMap["auto_rotate"] = model.AutoRotate
	}
	if model.Interval != nil {
		modelMap["interval"] = flex.IntValue(model.Interval)
	}
	if model.Unit != nil {
		modelMap["unit"] = model.Unit
	}
	return modelMap, nil
}

func resourceIbmSmServiceCredentialsSecretSourceServiceToMap(sourceService *secretsmanagerv2.ServiceCredentialsSecretSourceService) (map[string]interface{}, error) {
	mainModelMap := make(map[string]interface{})
	if sourceService.Instance != nil {
		instanceMap := make(map[string]interface{})
		instanceModel := sourceService.Instance
		if instanceModel.Crn != nil {
			instanceMap["crn"] = instanceModel.Crn
		}
		mainModelMap["instance"] = []map[string]interface{}{instanceMap}
	}

	if sourceService.Role != nil {
		roleMap := make(map[string]interface{})
		roleModel := sourceService.Role
		if roleModel.Crn != nil {
			roleMap["crn"] = roleModel.Crn
		}
		mainModelMap["role"] = []map[string]interface{}{roleMap}
	}

	if sourceService.Iam != nil {
		iamMap := make(map[string]interface{})
		iamModel := sourceService.Iam

		// apikey
		if iamModel.Apikey != nil {
			iamApikeyMap := make(map[string]interface{})
			iamApikeyModel := iamModel.Apikey
			if iamApikeyModel.Name != nil {
				iamApikeyMap["name"] = iamApikeyModel.Name
			}
			if iamApikeyModel.Description != nil {
				iamApikeyMap["description"] = iamApikeyModel.Description
			}
			iamMap["apikey"] = []map[string]interface{}{iamApikeyMap}
		}

		// role
		if iamModel.Role != nil {
			iamRoleMap := make(map[string]interface{})
			iamRoleModel := iamModel.Role
			if iamRoleModel.Crn != nil {
				iamRoleMap["crn"] = iamRoleModel.Crn
			}
			iamMap["role"] = []map[string]interface{}{iamRoleMap}
		}

		// service id
		if iamModel.Serviceid != nil {
			iamServiceidMap := make(map[string]interface{})
			iamServiceidModel := iamModel.Serviceid
			if iamServiceidModel.Crn != nil {
				iamServiceidMap["crn"] = iamServiceidModel.Crn
			}
			iamMap["serviceid"] = []map[string]interface{}{iamServiceidMap}
		}

		mainModelMap["iam"] = []map[string]interface{}{iamMap}

	}

	if sourceService.ResourceKey != nil {
		resourceKeyMap := make(map[string]interface{})
		resourceKeyModel := sourceService.ResourceKey
		if resourceKeyModel.Crn != nil {
			resourceKeyMap["crn"] = resourceKeyModel.Crn
		}
		if resourceKeyModel.Name != nil {
			resourceKeyMap["name"] = resourceKeyModel.Name
		}
		mainModelMap["resource_key"] = []map[string]interface{}{resourceKeyMap}
	}

	if sourceService.Parameters != nil {
		parametersMap := sourceService.Parameters.GetProperties()
		for k, v := range parametersMap {
			parametersMap[k] = fmt.Sprint(v)
		}
		if sourceService.Parameters.ServiceidCrn != nil {
			parametersMap["serviceid_crn"] = sourceService.Parameters.ServiceidCrn
		}
		mainModelMap["parameters"] = parametersMap
	}

	return mainModelMap, nil
}

func resourceIbmSmServiceCredentialsSecretCredentialsToMap(credentials *secretsmanagerv2.ServiceCredentialsSecretCredentials) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	if credentials.IamApikeyDescription != nil {
		modelMap["iam_apikey_description"] = credentials.IamApikeyDescription
	}
	if credentials.Apikey != nil {
		modelMap["apikey"] = credentials.Apikey
	}
	if credentials.Endpoints != nil {
		modelMap["endpoints"] = credentials.Endpoints
	}
	if credentials.IamApikeyName != nil {
		modelMap["iam_apikey_name"] = credentials.IamApikeyName
	}
	if credentials.IamRoleCrn != nil {
		modelMap["iam_role_crn"] = credentials.IamRoleCrn
	}
	if credentials.IamServiceidCrn != nil {
		modelMap["iam_serviceid_crn"] = credentials.IamServiceidCrn
	}
	if credentials.ResourceInstanceID != nil {
		modelMap["resource_instance_id"] = credentials.ResourceInstanceID
	}
	if credentials.CosHmacKeys != nil {
		cosHmacKeys := [1]map[string]interface{}{}
		m := map[string]interface{}{}
		if credentials.CosHmacKeys.AccessKeyID != nil {
			m["access_key_id"] = credentials.CosHmacKeys.AccessKeyID
		}
		if credentials.CosHmacKeys.SecretAccessKey != nil {
			m["secret_access_key"] = credentials.CosHmacKeys.SecretAccessKey
		}
		cosHmacKeys[0] = m
		modelMap["cos_hmac_keys"] = cosHmacKeys
	}
	return modelMap, nil
}
