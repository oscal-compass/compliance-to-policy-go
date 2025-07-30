package subcommands

import (
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	testOptions := &Options{}

	tests := []struct {
		name      string
		ap        *oscalTypes.AssessmentPlan
		wantError string
	}{
		{
			name: "Success/HappyPath",
			ap: &oscalTypes.AssessmentPlan{
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Activities: &[]oscalTypes.Activity{
						{
							Title: "my-activity",
							UUID:  "example-3",
							Steps: &[]oscalTypes.Step{
								{
									Title: "my-step",
									UUID:  "example-4",
								},
							},
						},
					},
					Components: &[]oscalTypes.SystemComponent{
						{
							UUID:  "uuid-1",
							Title: "example",
						},
					},
				},
				AssessmentAssets: &oscalTypes.AssessmentAssets{
					Components: &[]oscalTypes.SystemComponent{
						{
							UUID:  "uuid-2",
							Title: "example-validation",
						},
					},
				},
			},
		},
		{
			name:      "Invalid/NoActivities",
			wantError: "no activities found in assessment plan \"My Plan\"",
			ap: &oscalTypes.AssessmentPlan{
				Metadata: oscalTypes.Metadata{
					Title: "My Plan",
				},
				LocalDefinitions: &oscalTypes.LocalDefinitions{
					Components: &[]oscalTypes.SystemComponent{
						{
							UUID:  "uuid-1",
							Title: "example",
						},
					},
				},
				AssessmentAssets: &oscalTypes.AssessmentAssets{
					Components: &[]oscalTypes.SystemComponent{
						{
							UUID:  "uuid-2",
							Title: "example-validation",
						},
					},
				},
			},
		},
		{
			name:      "Invalid/MissingComponent",
			wantError: "missing components in assessment plan \"My Plan\"",
			ap: &oscalTypes.AssessmentPlan{
				Metadata: oscalTypes.Metadata{
					Title: "My Plan",
				},
				LocalDefinitions: &oscalTypes.LocalDefinitions{},
				AssessmentAssets: &oscalTypes.AssessmentAssets{
					Components: &[]oscalTypes.SystemComponent{
						{
							UUID:  "uuid-2",
							Title: "example-validation",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Context(testOptions, test.ap)
			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}

}
