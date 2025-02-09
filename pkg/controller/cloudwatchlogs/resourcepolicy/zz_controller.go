/*
Copyright 2021 The Crossplane Authors.

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

// Code generated by ack-generate. DO NOT EDIT.

package resourcepolicy

import (
	"context"

	svcapi "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	svcsdk "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	svcsdkapi "github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudwatchlogs/v1alpha1"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

const (
	errUnexpectedObject = "managed resource is not an ResourcePolicy resource"

	errCreateSession = "cannot create a new session"
	errCreate        = "cannot create ResourcePolicy in AWS"
	errUpdate        = "cannot update ResourcePolicy in AWS"
	errDescribe      = "failed to describe ResourcePolicy"
	errDelete        = "failed to delete ResourcePolicy"
)

type connector struct {
	kube client.Client
	opts []option
}

func (c *connector) Connect(ctx context.Context, cr *svcapitypes.ResourcePolicy) (managed.TypedExternalClient[*svcapitypes.ResourcePolicy], error) {
	sess, err := connectaws.GetConfigV1(ctx, c.kube, cr, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return newExternal(c.kube, svcapi.New(sess), c.opts), nil
}

func (e *external) Observe(ctx context.Context, cr *svcapitypes.ResourcePolicy) (managed.ExternalObservation, error) {
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	input := GenerateDescribeResourcePoliciesInput(cr)
	if err := e.preObserve(ctx, cr, input); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "pre-observe failed")
	}
	resp, err := e.client.DescribeResourcePoliciesWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
	}
	resp = e.filterList(cr, resp)
	if len(resp.ResourcePolicies) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err := e.lateInitialize(&cr.Spec.ForProvider, resp); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "late-init failed")
	}
	GenerateResourcePolicy(resp).Status.AtProvider.DeepCopyInto(&cr.Status.AtProvider)
	upToDate := true
	diff := ""
	if !meta.WasDeleted(cr) { // There is no need to run isUpToDate if the resource is deleted
		upToDate, diff, err = e.isUpToDate(ctx, cr, resp)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "isUpToDate check failed")
		}
	}
	return e.postObserve(ctx, cr, resp, managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		Diff:                    diff,
		ResourceLateInitialized: !cmp.Equal(&cr.Spec.ForProvider, currentSpec),
	}, nil)
}

func (e *external) Create(ctx context.Context, cr *svcapitypes.ResourcePolicy) (managed.ExternalCreation, error) {
	cr.Status.SetConditions(xpv1.Creating())
	input := GeneratePutResourcePolicyInput(cr)
	if err := e.preCreate(ctx, cr, input); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "pre-create failed")
	}
	resp, err := e.client.PutResourcePolicyWithContext(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	if resp.ResourcePolicy.LastUpdatedTime != nil {
		cr.Status.AtProvider.LastUpdatedTime = resp.ResourcePolicy.LastUpdatedTime
	} else {
		cr.Status.AtProvider.LastUpdatedTime = nil
	}
	if resp.ResourcePolicy.PolicyDocument != nil {
		cr.Spec.ForProvider.PolicyDocument = resp.ResourcePolicy.PolicyDocument
	} else {
		cr.Spec.ForProvider.PolicyDocument = nil
	}
	if resp.ResourcePolicy.PolicyName != nil {
		cr.Status.AtProvider.PolicyName = resp.ResourcePolicy.PolicyName
	} else {
		cr.Status.AtProvider.PolicyName = nil
	}

	return e.postCreate(ctx, cr, resp, managed.ExternalCreation{}, err)
}

func (e *external) Update(ctx context.Context, cr *svcapitypes.ResourcePolicy) (managed.ExternalUpdate, error) {
	return e.update(ctx, cr)

}

func (e *external) Delete(ctx context.Context, cr *svcapitypes.ResourcePolicy) (managed.ExternalDelete, error) {
	cr.Status.SetConditions(xpv1.Deleting())
	input := GenerateDeleteResourcePolicyInput(cr)
	ignore, err := e.preDelete(ctx, cr, input)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "pre-delete failed")
	}
	if ignore {
		return managed.ExternalDelete{}, nil
	}
	resp, err := e.client.DeleteResourcePolicyWithContext(ctx, input)
	return e.postDelete(ctx, cr, resp, errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDelete))
}

func (e *external) Disconnect(ctx context.Context) error {
	// Unimplemented, required by newer versions of crossplane-runtime
	return nil
}

type option func(*external)

func newExternal(kube client.Client, client svcsdkapi.CloudWatchLogsAPI, opts []option) *external {
	e := &external{
		kube:           kube,
		client:         client,
		preObserve:     nopPreObserve,
		postObserve:    nopPostObserve,
		lateInitialize: nopLateInitialize,
		isUpToDate:     alwaysUpToDate,
		filterList:     nopFilterList,
		preCreate:      nopPreCreate,
		postCreate:     nopPostCreate,
		preDelete:      nopPreDelete,
		postDelete:     nopPostDelete,
		update:         nopUpdate,
	}
	for _, f := range opts {
		f(e)
	}
	return e
}

type external struct {
	kube           client.Client
	client         svcsdkapi.CloudWatchLogsAPI
	preObserve     func(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.DescribeResourcePoliciesInput) error
	postObserve    func(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.DescribeResourcePoliciesOutput, managed.ExternalObservation, error) (managed.ExternalObservation, error)
	filterList     func(*svcapitypes.ResourcePolicy, *svcsdk.DescribeResourcePoliciesOutput) *svcsdk.DescribeResourcePoliciesOutput
	lateInitialize func(*svcapitypes.ResourcePolicyParameters, *svcsdk.DescribeResourcePoliciesOutput) error
	isUpToDate     func(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.DescribeResourcePoliciesOutput) (bool, string, error)
	preCreate      func(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.PutResourcePolicyInput) error
	postCreate     func(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.PutResourcePolicyOutput, managed.ExternalCreation, error) (managed.ExternalCreation, error)
	preDelete      func(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.DeleteResourcePolicyInput) (bool, error)
	postDelete     func(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.DeleteResourcePolicyOutput, error) (managed.ExternalDelete, error)
	update         func(context.Context, *svcapitypes.ResourcePolicy) (managed.ExternalUpdate, error)
}

func nopPreObserve(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.DescribeResourcePoliciesInput) error {
	return nil
}
func nopPostObserve(_ context.Context, _ *svcapitypes.ResourcePolicy, _ *svcsdk.DescribeResourcePoliciesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}
func nopFilterList(_ *svcapitypes.ResourcePolicy, list *svcsdk.DescribeResourcePoliciesOutput) *svcsdk.DescribeResourcePoliciesOutput {
	return list
}

func nopLateInitialize(*svcapitypes.ResourcePolicyParameters, *svcsdk.DescribeResourcePoliciesOutput) error {
	return nil
}
func alwaysUpToDate(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.DescribeResourcePoliciesOutput) (bool, string, error) {
	return true, "", nil
}

func nopPreCreate(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.PutResourcePolicyInput) error {
	return nil
}
func nopPostCreate(_ context.Context, _ *svcapitypes.ResourcePolicy, _ *svcsdk.PutResourcePolicyOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}
func nopPreDelete(context.Context, *svcapitypes.ResourcePolicy, *svcsdk.DeleteResourcePolicyInput) (bool, error) {
	return false, nil
}
func nopPostDelete(_ context.Context, _ *svcapitypes.ResourcePolicy, _ *svcsdk.DeleteResourcePolicyOutput, err error) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, err
}
func nopUpdate(context.Context, *svcapitypes.ResourcePolicy) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}
