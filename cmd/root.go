package cmd

import (
	"context"
	"fmt"

	"github.com/google/go-github/v50/github"
	"github.com/rajatjindal/fermyon-cloud-preview/pkg/cloud"
	"github.com/sethvargo/go-githubactions"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fermyon-cloud-preview",
	Short: "fermyon-cloud-preview is a Github Action for deploying preview for pull requests",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		//TODO: add option to override
		fermyonClient, err := cloud.NewClient(cloud.ProductionCloudLink)
		if err != nil {
			logrus.WithError(err).Fatalf("failed to create fermyon cloud client")
		}

		// if number of apps > 5, delete the oldest PR (behind a flag)
		allApps, err := fermyonClient.GetAllApps()
		if err != nil {
			logrus.WithError(err).Error("failed to get apps count")
		}

		appName, err := cloud.GetAppNameFromSpinToml()
		if err != nil {
			logrus.WithError(err).Fatalf("failed to find app name from spin.toml")
		}

		if len(allApps) > 5 {
			if githubactions.GetInput("overwrite_old_preview") != "true" {
				logrus.WithField("deployed_apps_count", len(allApps)).Fatalf("apps quota exceeded in cloud. set 'overwrite_old_preview' to true to overwrite old previews")
			}

			prToUndeploy, err := GetOldestPreviewPR(ctx)
			if err != nil {
				logrus.WithError(err).Fatalf("failed to find oldest pr to undeploy")
			}

			previewAppName := fmt.Sprintf("%s-pr-%d", appName, prToUndeploy)
			err = fermyonClient.DeleteAppByName(previewAppName)
			if err != nil {
				logrus.WithError(err).WithField("app_name", previewAppName).Fatalf("failed to undeploy preview app")
			}
		}

		metadata, err := fermyonClient.Deploy(appName)
		if err != nil {
			logrus.WithError(err).Error("preview deployment failed")
		}

		comment := fmt.Sprintf("your app preview is available at %s", metadata.Base)
		err = UpdateComment(ctx, 10, comment)
		if err != nil {
			logrus.WithError(err).Error("update comment on pr failed")
		}
	},
}

func GetOldestPreviewPR(ctx context.Context) (int, error) {
	client := github.NewClient(nil)

	ghContext, _ := githubactions.Context()
	owner, repo := ghContext.Repo()

	query := fmt.Sprintf("org:%s repo:%s is:open is:pr", owner, repo)
	issues, _, err := client.Search.Issues(ctx, query, &github.SearchOptions{Sort: "updated", Order: "asc"})
	if err != nil {
		return 0, err
	}

	if len(issues.Issues) == 0 {
		return 0, fmt.Errorf("no old previews found to undeploy")
	}

	return issues.Issues[0].GetNumber(), nil
}

func UpdateComment(ctx context.Context, prNumber int, comment string) error {
	client := github.NewClient(nil)

	ghContext, _ := githubactions.Context()
	owner, repo := ghContext.Repo()

	ghComment := &github.IssueComment{
		Body: &comment,
	}

	_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, ghComment)

	return err
}

func Execute() error {
	return rootCmd.Execute()
}
