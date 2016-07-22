package github

import (
	"github.com/google/go-github/github"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGithubTeamRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubTeamRepositoryCreate,
		Read:   resourceGithubTeamRepositoryRead,
		Update: resourceGithubTeamRepositoryUpdate,
		Delete: resourceGithubTeamRepositoryDelete,

		Schema: map[string]*schema.Schema{
			"team_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"permission": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "pull",
				ValidateFunc: validateValueFunc([]string{"pull", "push", "admin"}),
			},
		},
	}
}

func resourceGithubTeamRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).client
	t := d.Get("team_id").(string)
	r := d.Get("repository").(string)
	p := d.Get("permission").(string)

	_, err := client.Organizations.AddTeamRepo(toGithubID(t), meta.(*Organization).name, r,
		&github.OrganizationAddTeamRepoOptions{Permission: p})

	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(&t, &r))

	return resourceGithubTeamRepositoryRead(d, meta)
}

func resourceGithubTeamRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).client
	t := d.Get("team_id").(string)
	r := d.Get("repository").(string)

	repo, _, repoErr := client.Organizations.IsTeamRepo(toGithubID(t), meta.(*Organization).name, r)

	if repoErr != nil {
		d.SetId("")
		return nil
	}

	repositoryName := repo.Name

	d.Set("team_id", t)
	d.Set("repository", repositoryName)

	permName, permErr := getRepoPermission(repo.Permissions)

	if permErr != nil {
		return permErr
	}

	d.Set("permission", permName)

	return nil
}

func resourceGithubTeamRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).client
	t := d.Get("team_id").(string)
	r := d.Get("repository").(string)
	p := d.Get("permission").(string)

	// the go-github library's AddTeamRepo method uses the add/update endpoint from Github API
	_, err := client.Organizations.AddTeamRepo(toGithubID(t), meta.(*Organization).name, r,
		&github.OrganizationAddTeamRepoOptions{Permission: p})

	if err != nil {
		return err
	}
	return resourceGithubTeamRepositoryRead(d, meta)
}

func resourceGithubTeamRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).client
	t := d.Get("team_id").(string)
	r := d.Get("repository").(string)

	_, err := client.Organizations.RemoveTeamRepo(toGithubID(t), meta.(*Organization).name, r)

	return err
}
