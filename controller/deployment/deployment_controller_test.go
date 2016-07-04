/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package deployment

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/api/unversioned"
	exp "k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	"k8s.io/kubernetes/pkg/client/record"
	"k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util"
	"k8s.io/kubernetes/pkg/util/intstr"
)

var (
	alwaysReady = func() bool { return true }
	noTimestamp = unversioned.Time{}
)

func rs(name string, replicas int, selector map[string]string, timestamp unversioned.Time) *exp.ReplicaSet {
	return &exp.ReplicaSet{
		ObjectMeta: api.ObjectMeta{
			Name:              name,
			CreationTimestamp: timestamp,
			Namespace:         api.NamespaceDefault,
		},
		Spec: exp.ReplicaSetSpec{
			Replicas: int32(replicas),
			Selector: &unversioned.LabelSelector{MatchLabels: selector},
			Template: api.PodTemplateSpec{},
		},
	}
}

func newRSWithStatus(name string, specReplicas, statusReplicas int, selector map[string]string) *exp.ReplicaSet {
	rs := rs(name, specReplicas, selector, noTimestamp)
	rs.Status = exp.ReplicaSetStatus{
		Replicas: int32(statusReplicas),
	}
	return rs
}

func deployment(name string, replicas int, maxSurge, maxUnavailable intstr.IntOrString, selector map[string]string) exp.Deployment {
	return exp.Deployment{
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: api.NamespaceDefault,
		},
		Spec: exp.DeploymentSpec{
			Replicas: int32(replicas),
			Selector: &unversioned.LabelSelector{MatchLabels: selector},
			Strategy: exp.DeploymentStrategy{
				Type: exp.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &exp.RollingUpdateDeployment{
					MaxSurge:       maxSurge,
					MaxUnavailable: maxUnavailable,
				},
			},
		},
	}
}

func newDeployment(replicas int, revisionHistoryLimit *int) *exp.Deployment {
	var v *int32
	if revisionHistoryLimit != nil {
		v = new(int32)
		*v = int32(*revisionHistoryLimit)
	}
	d := exp.Deployment{
		TypeMeta: unversioned.TypeMeta{APIVersion: testapi.Default.GroupVersion().String()},
		ObjectMeta: api.ObjectMeta{
			UID:             util.NewUUID(),
			Name:            "foobar",
			Namespace:       api.NamespaceDefault,
			ResourceVersion: "18",
		},
		Spec: exp.DeploymentSpec{
			Strategy: exp.DeploymentStrategy{
				Type:          exp.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &exp.RollingUpdateDeployment{},
			},
			Replicas: int32(replicas),
			Selector: &unversioned.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: map[string]string{
						"name": "foo",
						"type": "production",
					},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Image: "foo/bar",
						},
					},
				},
			},
			RevisionHistoryLimit: v,
		},
	}
	return &d
}

// TODO: Consolidate all deployment helpers into one.
func newDeploymentEnhanced(replicas int, maxSurge intstr.IntOrString) *exp.Deployment {
	d := newDeployment(replicas, nil)
	d.Spec.Strategy.RollingUpdate.MaxSurge = maxSurge
	return d
}

func newReplicaSet(d *exp.Deployment, name string, replicas int) *exp.ReplicaSet {
	return &exp.ReplicaSet{
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: api.NamespaceDefault,
		},
		Spec: exp.ReplicaSetSpec{
			Replicas: int32(replicas),
			Template: d.Spec.Template,
		},
	}
}

// TestScale tests proportional scaling of deployments. Note that fenceposts for
// rolling out (maxUnavailable, maxSurge) have no meaning for simple scaling other
// than recording maxSurge as part of the max-replicas annotation that is taken
// into account in the next scale event (max-replicas is used for calculating the
// proportion of a replica set).
func TestScale(t *testing.T) {
	newTimestamp := unversioned.Date(2016, 5, 20, 2, 0, 0, 0, time.UTC)
	oldTimestamp := unversioned.Date(2016, 5, 20, 1, 0, 0, 0, time.UTC)
	olderTimestamp := unversioned.Date(2016, 5, 20, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		deployment    *exp.Deployment
		oldDeployment *exp.Deployment

		newRS  *exp.ReplicaSet
		oldRSs []*exp.ReplicaSet

		expectedNew *exp.ReplicaSet
		expectedOld []*exp.ReplicaSet

		desiredReplicasAnnotations map[string]int32
	}{
		{
			name:          "normal scaling event: 10 -> 12",
			deployment:    newDeployment(12, nil),
			oldDeployment: newDeployment(10, nil),

			newRS:  rs("foo-v1", 10, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{},

			expectedNew: rs("foo-v1", 12, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{},
		},
		{
			name:          "normal scaling event: 10 -> 5",
			deployment:    newDeployment(5, nil),
			oldDeployment: newDeployment(10, nil),

			newRS:  rs("foo-v1", 10, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{},

			expectedNew: rs("foo-v1", 5, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{},
		},
		{
			name:          "proportional scaling: 5 -> 10",
			deployment:    newDeployment(10, nil),
			oldDeployment: newDeployment(5, nil),

			newRS:  rs("foo-v2", 2, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v1", 3, nil, oldTimestamp)},

			expectedNew: rs("foo-v2", 4, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v1", 6, nil, oldTimestamp)},
		},
		{
			name:          "proportional scaling: 5 -> 3",
			deployment:    newDeployment(3, nil),
			oldDeployment: newDeployment(5, nil),

			newRS:  rs("foo-v2", 2, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v1", 3, nil, oldTimestamp)},

			expectedNew: rs("foo-v2", 1, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v1", 2, nil, oldTimestamp)},
		},
		{
			name:          "proportional scaling: 9 -> 4",
			deployment:    newDeployment(4, nil),
			oldDeployment: newDeployment(9, nil),

			newRS:  rs("foo-v2", 8, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v1", 1, nil, oldTimestamp)},

			expectedNew: rs("foo-v2", 4, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v1", 0, nil, oldTimestamp)},
		},
		{
			name:          "proportional scaling: 7 -> 10",
			deployment:    newDeployment(10, nil),
			oldDeployment: newDeployment(7, nil),

			newRS:  rs("foo-v3", 2, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v2", 3, nil, oldTimestamp), rs("foo-v1", 2, nil, olderTimestamp)},

			expectedNew: rs("foo-v3", 3, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v2", 4, nil, oldTimestamp), rs("foo-v1", 3, nil, olderTimestamp)},
		},
		{
			name:          "proportional scaling: 13 -> 8",
			deployment:    newDeployment(8, nil),
			oldDeployment: newDeployment(13, nil),

			newRS:  rs("foo-v3", 2, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v2", 8, nil, oldTimestamp), rs("foo-v1", 3, nil, olderTimestamp)},

			expectedNew: rs("foo-v3", 1, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v2", 5, nil, oldTimestamp), rs("foo-v1", 2, nil, olderTimestamp)},
		},
		// Scales up the new replica set.
		{
			name:          "leftover distribution: 3 -> 4",
			deployment:    newDeployment(4, nil),
			oldDeployment: newDeployment(3, nil),

			newRS:  rs("foo-v3", 1, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v2", 1, nil, oldTimestamp), rs("foo-v1", 1, nil, olderTimestamp)},

			expectedNew: rs("foo-v3", 2, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v2", 1, nil, oldTimestamp), rs("foo-v1", 1, nil, olderTimestamp)},
		},
		// Scales down the older replica set.
		{
			name:          "leftover distribution: 3 -> 2",
			deployment:    newDeployment(2, nil),
			oldDeployment: newDeployment(3, nil),

			newRS:  rs("foo-v3", 1, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v2", 1, nil, oldTimestamp), rs("foo-v1", 1, nil, olderTimestamp)},

			expectedNew: rs("foo-v3", 1, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v2", 1, nil, oldTimestamp), rs("foo-v1", 0, nil, olderTimestamp)},
		},
		// Scales up the latest replica set first.
		{
			name:          "proportional scaling (no new rs): 4 -> 5",
			deployment:    newDeployment(5, nil),
			oldDeployment: newDeployment(4, nil),

			newRS:  nil,
			oldRSs: []*exp.ReplicaSet{rs("foo-v2", 2, nil, oldTimestamp), rs("foo-v1", 2, nil, olderTimestamp)},

			expectedNew: nil,
			expectedOld: []*exp.ReplicaSet{rs("foo-v2", 3, nil, oldTimestamp), rs("foo-v1", 2, nil, olderTimestamp)},
		},
		// Scales down to zero
		{
			name:          "proportional scaling: 6 -> 0",
			deployment:    newDeployment(0, nil),
			oldDeployment: newDeployment(6, nil),

			newRS:  rs("foo-v3", 3, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v2", 2, nil, oldTimestamp), rs("foo-v1", 1, nil, olderTimestamp)},

			expectedNew: rs("foo-v3", 0, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v2", 0, nil, oldTimestamp), rs("foo-v1", 0, nil, olderTimestamp)},
		},
		// Scales up from zero
		{
			name:          "proportional scaling: 0 -> 6",
			deployment:    newDeployment(6, nil),
			oldDeployment: newDeployment(0, nil),

			newRS:  rs("foo-v3", 0, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v2", 0, nil, oldTimestamp), rs("foo-v1", 0, nil, olderTimestamp)},

			expectedNew: rs("foo-v3", 6, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v2", 0, nil, oldTimestamp), rs("foo-v1", 0, nil, olderTimestamp)},
		},
		// Scenario: deployment.spec.replicas == 3 ( foo-v1.spec.replicas == foo-v2.spec.replicas == foo-v3.spec.replicas == 1 )
		// Deployment is scaled to 5. foo-v3.spec.replicas and foo-v2.spec.replicas should increment by 1 but foo-v2 fails to
		// update.
		{
			name:          "failed rs update",
			deployment:    newDeployment(5, nil),
			oldDeployment: newDeployment(5, nil),

			newRS:  rs("foo-v3", 2, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v2", 1, nil, oldTimestamp), rs("foo-v1", 1, nil, olderTimestamp)},

			expectedNew: rs("foo-v3", 2, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v2", 2, nil, oldTimestamp), rs("foo-v1", 1, nil, olderTimestamp)},

			desiredReplicasAnnotations: map[string]int32{"foo-v2": int32(3)},
		},
		{
			name:          "deployment with surge pods",
			deployment:    newDeploymentEnhanced(20, intstr.FromInt(2)),
			oldDeployment: newDeploymentEnhanced(10, intstr.FromInt(2)),

			newRS:  rs("foo-v2", 6, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v1", 6, nil, oldTimestamp)},

			expectedNew: rs("foo-v2", 11, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v1", 11, nil, oldTimestamp)},
		},
		{
			name:          "change both surge and size",
			deployment:    newDeploymentEnhanced(50, intstr.FromInt(6)),
			oldDeployment: newDeploymentEnhanced(10, intstr.FromInt(3)),

			newRS:  rs("foo-v2", 5, nil, newTimestamp),
			oldRSs: []*exp.ReplicaSet{rs("foo-v1", 8, nil, oldTimestamp)},

			expectedNew: rs("foo-v2", 22, nil, newTimestamp),
			expectedOld: []*exp.ReplicaSet{rs("foo-v1", 34, nil, oldTimestamp)},
		},
	}

	for _, test := range tests {
		_ = olderTimestamp
		t.Log(test.name)
		fake := fake.Clientset{}
		dc := &DeploymentController{
			client:        &fake,
			eventRecorder: &record.FakeRecorder{},
		}

		if test.newRS != nil {
			desiredReplicas := test.oldDeployment.Spec.Replicas
			if desired, ok := test.desiredReplicasAnnotations[test.newRS.Name]; ok {
				desiredReplicas = desired
			}
			setReplicasAnnotations(test.newRS, desiredReplicas, desiredReplicas+maxSurge(*test.oldDeployment))
		}
		for i := range test.oldRSs {
			rs := test.oldRSs[i]
			if rs == nil {
				continue
			}
			desiredReplicas := test.oldDeployment.Spec.Replicas
			if desired, ok := test.desiredReplicasAnnotations[rs.Name]; ok {
				desiredReplicas = desired
			}
			setReplicasAnnotations(rs, desiredReplicas, desiredReplicas+maxSurge(*test.oldDeployment))
		}

		if err := dc.scale(test.deployment, test.newRS, test.oldRSs); err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
			continue
		}
		if test.expectedNew != nil && test.newRS != nil && test.expectedNew.Spec.Replicas != test.newRS.Spec.Replicas {
			t.Errorf("%s: expected new replicas: %d, got: %d", test.name, test.expectedNew.Spec.Replicas, test.newRS.Spec.Replicas)
			continue
		}
		if len(test.expectedOld) != len(test.oldRSs) {
			t.Errorf("%s: expected %d old replica sets, got %d", test.name, len(test.expectedOld), len(test.oldRSs))
			continue
		}
		for n := range test.oldRSs {
			rs := test.oldRSs[n]
			exp := test.expectedOld[n]
			if exp.Spec.Replicas != rs.Spec.Replicas {
				t.Errorf("%s: expected old (%s) replicas: %d, got: %d", test.name, rs.Name, exp.Spec.Replicas, rs.Spec.Replicas)
			}
		}
	}
}

func TestDeploymentController_reconcileNewReplicaSet(t *testing.T) {
	tests := []struct {
		deploymentReplicas  int
		maxSurge            intstr.IntOrString
		oldReplicas         int
		newReplicas         int
		scaleExpected       bool
		expectedNewReplicas int
	}{
		{
			// Should not scale up.
			deploymentReplicas: 10,
			maxSurge:           intstr.FromInt(0),
			oldReplicas:        10,
			newReplicas:        0,
			scaleExpected:      false,
		},
		{
			deploymentReplicas:  10,
			maxSurge:            intstr.FromInt(2),
			oldReplicas:         10,
			newReplicas:         0,
			scaleExpected:       true,
			expectedNewReplicas: 2,
		},
		{
			deploymentReplicas:  10,
			maxSurge:            intstr.FromInt(2),
			oldReplicas:         5,
			newReplicas:         0,
			scaleExpected:       true,
			expectedNewReplicas: 7,
		},
		{
			deploymentReplicas: 10,
			maxSurge:           intstr.FromInt(2),
			oldReplicas:        10,
			newReplicas:        2,
			scaleExpected:      false,
		},
		{
			// Should scale down.
			deploymentReplicas:  10,
			maxSurge:            intstr.FromInt(2),
			oldReplicas:         2,
			newReplicas:         11,
			scaleExpected:       true,
			expectedNewReplicas: 10,
		},
	}

	for i, test := range tests {
		t.Logf("executing scenario %d", i)
		newRS := rs("foo-v2", test.newReplicas, nil, noTimestamp)
		oldRS := rs("foo-v2", test.oldReplicas, nil, noTimestamp)
		allRSs := []*exp.ReplicaSet{newRS, oldRS}
		deployment := deployment("foo", test.deploymentReplicas, test.maxSurge, intstr.FromInt(0), nil)
		fake := fake.Clientset{}
		controller := &DeploymentController{
			client:        &fake,
			eventRecorder: &record.FakeRecorder{},
		}
		scaled, err := controller.reconcileNewReplicaSet(allRSs, newRS, &deployment)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if !test.scaleExpected {
			if scaled || len(fake.Actions()) > 0 {
				t.Errorf("unexpected scaling: %v", fake.Actions())
			}
			continue
		}
		if test.scaleExpected && !scaled {
			t.Errorf("expected scaling to occur")
			continue
		}
		if len(fake.Actions()) != 1 {
			t.Errorf("expected 1 action during scale, got: %v", fake.Actions())
			continue
		}
		updated := fake.Actions()[0].(core.UpdateAction).GetObject().(*exp.ReplicaSet)
		if e, a := test.expectedNewReplicas, int(updated.Spec.Replicas); e != a {
			t.Errorf("expected update to %d replicas, got %d", e, a)
		}
	}
}

func TestDeploymentController_reconcileOldReplicaSets(t *testing.T) {
	tests := []struct {
		deploymentReplicas  int
		maxUnavailable      intstr.IntOrString
		oldReplicas         int
		newReplicas         int
		readyPodsFromOldRS  int
		readyPodsFromNewRS  int
		scaleExpected       bool
		expectedOldReplicas int
	}{
		{
			deploymentReplicas:  10,
			maxUnavailable:      intstr.FromInt(0),
			oldReplicas:         10,
			newReplicas:         0,
			readyPodsFromOldRS:  10,
			readyPodsFromNewRS:  0,
			scaleExpected:       true,
			expectedOldReplicas: 9,
		},
		{
			deploymentReplicas:  10,
			maxUnavailable:      intstr.FromInt(2),
			oldReplicas:         10,
			newReplicas:         0,
			readyPodsFromOldRS:  10,
			readyPodsFromNewRS:  0,
			scaleExpected:       true,
			expectedOldReplicas: 8,
		},
		{ // expect unhealthy replicas from old replica sets been cleaned up
			deploymentReplicas:  10,
			maxUnavailable:      intstr.FromInt(2),
			oldReplicas:         10,
			newReplicas:         0,
			readyPodsFromOldRS:  8,
			readyPodsFromNewRS:  0,
			scaleExpected:       true,
			expectedOldReplicas: 8,
		},
		{ // expect 1 unhealthy replica from old replica sets been cleaned up, and 1 ready pod been scaled down
			deploymentReplicas:  10,
			maxUnavailable:      intstr.FromInt(2),
			oldReplicas:         10,
			newReplicas:         0,
			readyPodsFromOldRS:  9,
			readyPodsFromNewRS:  0,
			scaleExpected:       true,
			expectedOldReplicas: 8,
		},
		{ // the unavailable pods from the newRS would not make us scale down old RSs in a further step
			deploymentReplicas: 10,
			maxUnavailable:     intstr.FromInt(2),
			oldReplicas:        8,
			newReplicas:        2,
			readyPodsFromOldRS: 8,
			readyPodsFromNewRS: 0,
			scaleExpected:      false,
		},
	}
	for i, test := range tests {
		t.Logf("executing scenario %d", i)

		newSelector := map[string]string{"foo": "new"}
		oldSelector := map[string]string{"foo": "old"}
		newRS := rs("foo-new", test.newReplicas, newSelector, noTimestamp)
		oldRS := rs("foo-old", test.oldReplicas, oldSelector, noTimestamp)
		oldRSs := []*exp.ReplicaSet{oldRS}
		allRSs := []*exp.ReplicaSet{oldRS, newRS}

		deployment := deployment("foo", test.deploymentReplicas, intstr.FromInt(0), test.maxUnavailable, newSelector)
		fakeClientset := fake.Clientset{}
		fakeClientset.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
			switch action.(type) {
			case core.ListAction:
				podList := &api.PodList{}
				for podIndex := 0; podIndex < test.readyPodsFromOldRS; podIndex++ {
					podList.Items = append(podList.Items, api.Pod{
						ObjectMeta: api.ObjectMeta{
							Name:   fmt.Sprintf("%s-oldReadyPod-%d", oldRS.Name, podIndex),
							Labels: oldSelector,
						},
						Status: api.PodStatus{
							Conditions: []api.PodCondition{
								{
									Type:   api.PodReady,
									Status: api.ConditionTrue,
								},
							},
						},
					})
				}
				for podIndex := 0; podIndex < test.oldReplicas-test.readyPodsFromOldRS; podIndex++ {
					podList.Items = append(podList.Items, api.Pod{
						ObjectMeta: api.ObjectMeta{
							Name:   fmt.Sprintf("%s-oldUnhealthyPod-%d", oldRS.Name, podIndex),
							Labels: oldSelector,
						},
						Status: api.PodStatus{
							Conditions: []api.PodCondition{
								{
									Type:   api.PodReady,
									Status: api.ConditionFalse,
								},
							},
						},
					})
				}
				for podIndex := 0; podIndex < test.readyPodsFromNewRS; podIndex++ {
					podList.Items = append(podList.Items, api.Pod{
						ObjectMeta: api.ObjectMeta{
							Name:   fmt.Sprintf("%s-newReadyPod-%d", oldRS.Name, podIndex),
							Labels: newSelector,
						},
						Status: api.PodStatus{
							Conditions: []api.PodCondition{
								{
									Type:   api.PodReady,
									Status: api.ConditionTrue,
								},
							},
						},
					})
				}
				for podIndex := 0; podIndex < test.oldReplicas-test.readyPodsFromOldRS; podIndex++ {
					podList.Items = append(podList.Items, api.Pod{
						ObjectMeta: api.ObjectMeta{
							Name:   fmt.Sprintf("%s-newUnhealthyPod-%d", oldRS.Name, podIndex),
							Labels: newSelector,
						},
						Status: api.PodStatus{
							Conditions: []api.PodCondition{
								{
									Type:   api.PodReady,
									Status: api.ConditionFalse,
								},
							},
						},
					})
				}
				return true, podList, nil
			}
			return false, nil, nil
		})
		controller := &DeploymentController{
			client:        &fakeClientset,
			eventRecorder: &record.FakeRecorder{},
		}

		scaled, err := controller.reconcileOldReplicaSets(allRSs, oldRSs, newRS, &deployment)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if !test.scaleExpected && scaled {
			t.Errorf("unexpected scaling: %v", fakeClientset.Actions())
		}
		if test.scaleExpected && !scaled {
			t.Errorf("expected scaling to occur")
			continue
		}
		continue
	}
}

func TestDeploymentController_cleanupUnhealthyReplicas(t *testing.T) {
	tests := []struct {
		oldReplicas          int
		readyPods            int
		unHealthyPods        int
		maxCleanupCount      int
		cleanupCountExpected int
	}{
		{
			oldReplicas:          10,
			readyPods:            8,
			unHealthyPods:        2,
			maxCleanupCount:      1,
			cleanupCountExpected: 1,
		},
		{
			oldReplicas:          10,
			readyPods:            8,
			unHealthyPods:        2,
			maxCleanupCount:      3,
			cleanupCountExpected: 2,
		},
		{
			oldReplicas:          10,
			readyPods:            8,
			unHealthyPods:        2,
			maxCleanupCount:      0,
			cleanupCountExpected: 0,
		},
		{
			oldReplicas:          10,
			readyPods:            10,
			unHealthyPods:        0,
			maxCleanupCount:      3,
			cleanupCountExpected: 0,
		},
	}

	for i, test := range tests {
		t.Logf("executing scenario %d", i)
		oldRS := rs("foo-v2", test.oldReplicas, nil, noTimestamp)
		oldRSs := []*exp.ReplicaSet{oldRS}
		deployment := deployment("foo", 10, intstr.FromInt(2), intstr.FromInt(2), nil)
		fakeClientset := fake.Clientset{}
		fakeClientset.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
			switch action.(type) {
			case core.ListAction:
				podList := &api.PodList{}
				for podIndex := 0; podIndex < test.readyPods; podIndex++ {
					podList.Items = append(podList.Items, api.Pod{
						ObjectMeta: api.ObjectMeta{
							Name: fmt.Sprintf("%s-readyPod-%d", oldRS.Name, podIndex),
						},
						Status: api.PodStatus{
							Conditions: []api.PodCondition{
								{
									Type:   api.PodReady,
									Status: api.ConditionTrue,
								},
							},
						},
					})
				}
				for podIndex := 0; podIndex < test.unHealthyPods; podIndex++ {
					podList.Items = append(podList.Items, api.Pod{
						ObjectMeta: api.ObjectMeta{
							Name: fmt.Sprintf("%s-unHealthyPod-%d", oldRS.Name, podIndex),
						},
						Status: api.PodStatus{
							Conditions: []api.PodCondition{
								{
									Type:   api.PodReady,
									Status: api.ConditionFalse,
								},
							},
						},
					})
				}
				return true, podList, nil
			}
			return false, nil, nil
		})

		controller := &DeploymentController{
			client:        &fakeClientset,
			eventRecorder: &record.FakeRecorder{},
		}
		_, cleanupCount, err := controller.cleanupUnhealthyReplicas(oldRSs, &deployment, 0, int32(test.maxCleanupCount))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if int(cleanupCount) != test.cleanupCountExpected {
			t.Errorf("expected %v unhealthy replicas been cleaned up, got %v", test.cleanupCountExpected, cleanupCount)
			continue
		}
	}
}

func TestDeploymentController_scaleDownOldReplicaSetsForRollingUpdate(t *testing.T) {
	tests := []struct {
		deploymentReplicas  int
		maxUnavailable      intstr.IntOrString
		readyPods           int
		oldReplicas         int
		scaleExpected       bool
		expectedOldReplicas int
	}{
		{
			deploymentReplicas:  10,
			maxUnavailable:      intstr.FromInt(0),
			readyPods:           10,
			oldReplicas:         10,
			scaleExpected:       true,
			expectedOldReplicas: 9,
		},
		{
			deploymentReplicas:  10,
			maxUnavailable:      intstr.FromInt(2),
			readyPods:           10,
			oldReplicas:         10,
			scaleExpected:       true,
			expectedOldReplicas: 8,
		},
		{
			deploymentReplicas: 10,
			maxUnavailable:     intstr.FromInt(2),
			readyPods:          8,
			oldReplicas:        10,
			scaleExpected:      false,
		},
		{
			deploymentReplicas: 10,
			maxUnavailable:     intstr.FromInt(2),
			readyPods:          10,
			oldReplicas:        0,
			scaleExpected:      false,
		},
		{
			deploymentReplicas: 10,
			maxUnavailable:     intstr.FromInt(2),
			readyPods:          1,
			oldReplicas:        10,
			scaleExpected:      false,
		},
	}

	for i, test := range tests {
		t.Logf("executing scenario %d", i)
		oldRS := rs("foo-v2", test.oldReplicas, nil, noTimestamp)
		allRSs := []*exp.ReplicaSet{oldRS}
		oldRSs := []*exp.ReplicaSet{oldRS}
		deployment := deployment("foo", test.deploymentReplicas, intstr.FromInt(0), test.maxUnavailable, map[string]string{"foo": "bar"})
		fakeClientset := fake.Clientset{}
		fakeClientset.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
			switch action.(type) {
			case core.ListAction:
				podList := &api.PodList{}
				for podIndex := 0; podIndex < test.readyPods; podIndex++ {
					podList.Items = append(podList.Items, api.Pod{
						ObjectMeta: api.ObjectMeta{
							Name:   fmt.Sprintf("%s-pod-%d", oldRS.Name, podIndex),
							Labels: map[string]string{"foo": "bar"},
						},
						Status: api.PodStatus{
							Conditions: []api.PodCondition{
								{
									Type:   api.PodReady,
									Status: api.ConditionTrue,
								},
							},
						},
					})
				}
				return true, podList, nil
			}
			return false, nil, nil
		})
		controller := &DeploymentController{
			client:        &fakeClientset,
			eventRecorder: &record.FakeRecorder{},
		}
		scaled, err := controller.scaleDownOldReplicaSetsForRollingUpdate(allRSs, oldRSs, &deployment)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if !test.scaleExpected {
			if scaled != 0 {
				t.Errorf("unexpected scaling: %v", fakeClientset.Actions())
			}
			continue
		}
		if test.scaleExpected && scaled == 0 {
			t.Errorf("expected scaling to occur; actions: %v", fakeClientset.Actions())
			continue
		}
		// There are both list and update actions logged, so extract the update
		// action for verification.
		var updateAction core.UpdateAction
		for _, action := range fakeClientset.Actions() {
			switch a := action.(type) {
			case core.UpdateAction:
				if updateAction != nil {
					t.Errorf("expected only 1 update action; had %v and found %v", updateAction, a)
				} else {
					updateAction = a
				}
			}
		}
		if updateAction == nil {
			t.Errorf("expected an update action")
			continue
		}
		updated := updateAction.GetObject().(*exp.ReplicaSet)
		if e, a := test.expectedOldReplicas, int(updated.Spec.Replicas); e != a {
			t.Errorf("expected update to %d replicas, got %d", e, a)
		}
	}
}

func TestDeploymentController_cleanupDeployment(t *testing.T) {
	selector := map[string]string{"foo": "bar"}

	tests := []struct {
		oldRSs               []*exp.ReplicaSet
		revisionHistoryLimit int
		expectedDeletions    int
	}{
		{
			oldRSs: []*exp.ReplicaSet{
				newRSWithStatus("foo-1", 0, 0, selector),
				newRSWithStatus("foo-2", 0, 0, selector),
				newRSWithStatus("foo-3", 0, 0, selector),
			},
			revisionHistoryLimit: 1,
			expectedDeletions:    2,
		},
		{
			// Only delete the replica set with Spec.Replicas = Status.Replicas = 0.
			oldRSs: []*exp.ReplicaSet{
				newRSWithStatus("foo-1", 0, 0, selector),
				newRSWithStatus("foo-2", 0, 1, selector),
				newRSWithStatus("foo-3", 1, 0, selector),
				newRSWithStatus("foo-4", 1, 1, selector),
			},
			revisionHistoryLimit: 0,
			expectedDeletions:    1,
		},

		{
			oldRSs: []*exp.ReplicaSet{
				newRSWithStatus("foo-1", 0, 0, selector),
				newRSWithStatus("foo-2", 0, 0, selector),
			},
			revisionHistoryLimit: 0,
			expectedDeletions:    2,
		},
		{
			oldRSs: []*exp.ReplicaSet{
				newRSWithStatus("foo-1", 1, 1, selector),
				newRSWithStatus("foo-2", 1, 1, selector),
			},
			revisionHistoryLimit: 0,
			expectedDeletions:    0,
		},
	}

	for i, test := range tests {
		fake := &fake.Clientset{}
		controller := NewDeploymentController(fake, controller.NoResyncPeriodFunc)

		controller.eventRecorder = &record.FakeRecorder{}
		controller.rsStoreSynced = alwaysReady
		controller.podStoreSynced = alwaysReady
		for _, rs := range test.oldRSs {
			controller.rsStore.Add(rs)
		}

		d := newDeployment(1, &tests[i].revisionHistoryLimit)
		controller.cleanupDeployment(test.oldRSs, d)

		gotDeletions := 0
		for _, action := range fake.Actions() {
			if "delete" == action.GetVerb() {
				gotDeletions++
			}
		}
		if gotDeletions != test.expectedDeletions {
			t.Errorf("expect %v old replica sets been deleted, but got %v", test.expectedDeletions, gotDeletions)
			continue
		}
	}
}

func getKey(d *exp.Deployment, t *testing.T) string {
	if key, err := controller.KeyFunc(d); err != nil {
		t.Errorf("Unexpected error getting key for deployment %v: %v", d.Name, err)
		return ""
	} else {
		return key
	}
}

type fixture struct {
	t *testing.T

	client *fake.Clientset
	// Objects to put in the store.
	dStore   []*exp.Deployment
	rsStore  []*exp.ReplicaSet
	podStore []*api.Pod

	// Actions expected to happen on the client. Objects from here are also
	// preloaded into NewSimpleFake.
	actions []core.Action
	objects []runtime.Object
}

func (f *fixture) expectUpdateDeploymentAction(d *exp.Deployment) {
	f.actions = append(f.actions, core.NewUpdateAction(unversioned.GroupVersionResource{Resource: "deployments"}, d.Namespace, d))
}

func (f *fixture) expectUpdateDeploymentStatusAction(d *exp.Deployment) {
	action := core.NewUpdateAction(unversioned.GroupVersionResource{Resource: "deployments"}, d.Namespace, d)
	action.Subresource = "status"
	f.actions = append(f.actions, action)
}

func (f *fixture) expectCreateRSAction(rs *exp.ReplicaSet) {
	f.actions = append(f.actions, core.NewCreateAction(unversioned.GroupVersionResource{Resource: "replicasets"}, rs.Namespace, rs))
}

func (f *fixture) expectUpdateRSAction(rs *exp.ReplicaSet) {
	f.actions = append(f.actions, core.NewUpdateAction(unversioned.GroupVersionResource{Resource: "replicasets"}, rs.Namespace, rs))
}

func (f *fixture) expectListPodAction(namespace string, opt api.ListOptions) {
	f.actions = append(f.actions, core.NewListAction(unversioned.GroupVersionResource{Resource: "pods"}, namespace, opt))
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	return f
}

func (f *fixture) run(deploymentName string) {
	f.client = fake.NewSimpleClientset(f.objects...)
	c := NewDeploymentController(f.client, controller.NoResyncPeriodFunc)
	c.eventRecorder = &record.FakeRecorder{}
	c.rsStoreSynced = alwaysReady
	c.podStoreSynced = alwaysReady
	for _, d := range f.dStore {
		c.dStore.Store.Add(d)
	}
	for _, rs := range f.rsStore {
		c.rsStore.Store.Add(rs)
	}
	for _, pod := range f.podStore {
		c.podStore.Indexer.Add(pod)
	}

	err := c.syncDeployment(deploymentName)
	if err != nil {
		f.t.Errorf("error syncing deployment: %v", err)
	}

	actions := f.client.Actions()
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		if !expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) {
			f.t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expectedAction, action)
			continue
		}
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}
}

func TestSyncDeploymentCreatesReplicaSet(t *testing.T) {
	f := newFixture(t)

	d := newDeployment(1, nil)
	f.dStore = append(f.dStore, d)
	f.objects = append(f.objects, d)

	rs := newReplicaSet(d, "deploymentrs-4186632231", 1)

	f.expectCreateRSAction(rs)
	f.expectUpdateDeploymentAction(d)
	f.expectUpdateDeploymentStatusAction(d)

	f.run(getKey(d, t))
}

// issue: https://github.com/kubernetes/kubernetes/issues/23218
func TestDeploymentController_dontSyncDeploymentsWithEmptyPodSelector(t *testing.T) {
	fake := &fake.Clientset{}
	controller := NewDeploymentController(fake, controller.NoResyncPeriodFunc)

	controller.eventRecorder = &record.FakeRecorder{}
	controller.rsStoreSynced = alwaysReady
	controller.podStoreSynced = alwaysReady

	d := newDeployment(1, nil)
	empty := unversioned.LabelSelector{}
	d.Spec.Selector = &empty
	controller.dStore.Store.Add(d)
	// We expect the deployment controller to not take action here since it's configuration
	// is invalid, even though no replicasets exist that match it's selector.
	controller.syncDeployment(fmt.Sprintf("%s/%s", d.ObjectMeta.Namespace, d.ObjectMeta.Name))
	if len(fake.Actions()) == 0 {
		return
	}
	for _, action := range fake.Actions() {
		t.Logf("unexpected action: %#v", action)
	}
	t.Errorf("expected deployment controller to not take action")
}
