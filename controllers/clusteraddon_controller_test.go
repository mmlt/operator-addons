package controllers

import (
	"github.com/google/go-cmp/cmp"
	"github.com/mmlt/operator-addons/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func Test_updateStatus(t *testing.T) {
	time1 := time.Unix(1, 0) // original time
	time2 := time.Unix(2, 0) // update time

	type args struct {
		clusterAddon *v1alpha1.ClusterAddon
		status       *v1alpha1.ClusterAddonStatus
	}
	tests := []struct {
		name   string
		args   args
		want   bool
		wantCA *v1alpha1.ClusterAddon
	}{
		{
			name: "Empty_Status_+_0_conditions",
			args: args{
				clusterAddon: &v1alpha1.ClusterAddon{
					Status: v1alpha1.ClusterAddonStatus{
						Conditions: nil,
						Synced:     metav1.ConditionFalse,
					},
				},
				status: &v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{},
					Synced:     metav1.ConditionFalse,
				},
			},
			want: false,
			wantCA: &v1alpha1.ClusterAddon{
				Status: v1alpha1.ClusterAddonStatus{
					Conditions: nil,
					Synced:     metav1.ConditionFalse,
				},
			},
		},
		{
			name: "Empty_Status_+_1_condition",
			args: args{
				clusterAddon: &v1alpha1.ClusterAddon{
					Status: v1alpha1.ClusterAddonStatus{
						Conditions: nil,
						Synced:     metav1.ConditionFalse,
					},
				},
				status: &v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time1},
							Reason:             "a",
							Message:            "b",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
			want: true,
			wantCA: &v1alpha1.ClusterAddon{
				Status: v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time2},
							Reason:             "a",
							Message:            "b",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
		},
		{
			name: "Empty_Status_+_true_true_conditions_of_same_type",
			args: args{
				clusterAddon: &v1alpha1.ClusterAddon{
					Status: v1alpha1.ClusterAddonStatus{
						Conditions: nil,
						Synced:     metav1.ConditionFalse,
					},
				},
				status: &v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time1},
							Reason:             "a",
							Message:            "b",
						},
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time1},
							Reason:             "c",
							Message:            "d",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
			want: true,
			wantCA: &v1alpha1.ClusterAddon{
				Status: v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time2},
							Reason:             "a",
							Message:            "b, d",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
		},
		{
			name: "Empty_Status_+_true_false_conditions_of_same_type",
			args: args{
				clusterAddon: &v1alpha1.ClusterAddon{
					Status: v1alpha1.ClusterAddonStatus{
						Conditions: nil,
						Synced:     metav1.ConditionFalse,
					},
				},
				status: &v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time1},
							Reason:             "a",
							Message:            "b",
						},
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time1},
							Reason:             "c",
							Message:            "d",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
			want: true,
			wantCA: &v1alpha1.ClusterAddon{
				Status: v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time2},
							Reason:             "a",
							Message:            "b, d",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
		},
		{
			name: "false_condition_Status_+_true_condition_of_same_type",
			args: args{
				clusterAddon: &v1alpha1.ClusterAddon{
					Status: v1alpha1.ClusterAddonStatus{
						Conditions: []v1alpha1.ClusterAddonCondition{
							v1alpha1.ClusterAddonCondition{
								Type:               v1alpha1.ClusterAddonSourceOk,
								Status:             metav1.ConditionFalse,
								LastTransitionTime: metav1.Time{Time: time1},
								Reason:             "a",
								Message:            "b",
							},
						},
						Synced: metav1.ConditionFalse,
					},
				},
				status: &v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time2},
							Reason:             "c",
							Message:            "d",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
			want: true,
			wantCA: &v1alpha1.ClusterAddon{
				Status: v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time2},
							Reason:             "c",
							Message:            "d",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
		},
		{
			name: "false_condition_Status_+_false_condition_of_same_type",
			args: args{
				clusterAddon: &v1alpha1.ClusterAddon{
					Status: v1alpha1.ClusterAddonStatus{
						Conditions: []v1alpha1.ClusterAddonCondition{
							v1alpha1.ClusterAddonCondition{
								Type:               v1alpha1.ClusterAddonSourceOk,
								Status:             metav1.ConditionFalse,
								LastTransitionTime: metav1.Time{Time: time1},
								Reason:             "a",
								Message:            "b",
							},
						},
						Synced: metav1.ConditionFalse,
					},
				},
				status: &v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time2},
							Reason:             "c",
							Message:            "d",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
			want: true,
			wantCA: &v1alpha1.ClusterAddon{
				Status: v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Time{Time: time1}, // time isn't updated as status stays the same
							Reason:             "c",                      // reason and message are updated to last known values
							Message:            "d",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
		},
		/*		{
				name: "true_condition_Status_+_no_conditions",
				args: args{
					clusterAddon: &v1alpha1.ClusterAddon{
						Status: v1alpha1.ClusterAddonStatus{
							Conditions: []v1alpha1.ClusterAddonCondition{
								v1alpha1.ClusterAddonCondition{
									Type:               v1alpha1.ClusterAddonSourceOk,
									Status:             metav1.ConditionTrue,
									LastTransitionTime: metav1.Time{Time: time1},
									Reason:             "a",
									Message:            "b",
								},
							},
							Synced: metav1.ConditionFalse,
						},
					},
					status: &v1alpha1.ClusterAddonStatus{
						Conditions: []v1alpha1.ClusterAddonCondition{},
						Synced:     metav1.ConditionFalse,
					},
				},
				want: true,
				wantCA: &v1alpha1.ClusterAddon{
					Status: v1alpha1.ClusterAddonStatus{
						Conditions: []v1alpha1.ClusterAddonCondition{
							v1alpha1.ClusterAddonCondition{
								Type:               v1alpha1.ClusterAddonSourceOk,
								Status:             metav1.ConditionUnknown, // status becomes unknown
								LastTransitionTime: metav1.Time{Time: time2},
								Reason:             "",
								Message:            "",
							},
						},
						Synced: metav1.ConditionFalse,
					},
				},
			},*/
		{
			name: "unknown_condition_Status_+_no_conditions",
			args: args{
				clusterAddon: &v1alpha1.ClusterAddon{
					Status: v1alpha1.ClusterAddonStatus{
						Conditions: []v1alpha1.ClusterAddonCondition{
							v1alpha1.ClusterAddonCondition{
								Type:               v1alpha1.ClusterAddonSourceOk,
								Status:             metav1.ConditionUnknown,
								LastTransitionTime: metav1.Time{Time: time1},
								Reason:             "",
								Message:            "",
							},
						},
						Synced: metav1.ConditionFalse,
					},
				},
				status: &v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{},
					Synced:     metav1.ConditionFalse,
				},
			},
			want: false,
			wantCA: &v1alpha1.ClusterAddon{
				Status: v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSourceOk,
							Status:             metav1.ConditionUnknown,
							LastTransitionTime: metav1.Time{Time: time1},
							Reason:             "",
							Message:            "",
						},
					},
					Synced: metav1.ConditionFalse,
				},
			},
		},
		{
			name: "0_condition_Status_+_1_synced_condition",
			args: args{
				clusterAddon: &v1alpha1.ClusterAddon{
					Status: v1alpha1.ClusterAddonStatus{
						Conditions: []v1alpha1.ClusterAddonCondition{},
						Synced:     metav1.ConditionFalse,
					},
				},
				status: &v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:   v1alpha1.ClusterAddonSynced,
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			want: true,
			wantCA: &v1alpha1.ClusterAddon{
				Status: v1alpha1.ClusterAddonStatus{
					Conditions: []v1alpha1.ClusterAddonCondition{
						v1alpha1.ClusterAddonCondition{
							Type:               v1alpha1.ClusterAddonSynced,
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: time2},
							Reason:             "",
							Message:            "",
						},
					},
					Synced: metav1.ConditionTrue, // ClusterAddonSynced Status is copied here.
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateStatus(tt.args.clusterAddon, tt.args.status, time2)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}

			if !cmp.Equal(tt.args.clusterAddon, tt.wantCA) {
				t.Errorf("diff (- = got, + = want) %s", cmp.Diff(tt.args.clusterAddon, tt.wantCA))
			}
		})
	}
}
