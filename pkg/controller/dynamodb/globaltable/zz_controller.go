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

package globaltable

import (
	"context"

	svcapi "github.com/aws/aws-sdk-go/service/dynamodb"
	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	svcsdkapi "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dynamodb/v1alpha1"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

const (
	errUnexpectedObject = "managed resource is not an GlobalTable resource"

	errCreateSession = "cannot create a new session"
	errCreate        = "cannot create GlobalTable in AWS"
	errUpdate        = "cannot update GlobalTable in AWS"
	errDescribe      = "failed to describe GlobalTable"
	errDelete        = "failed to delete GlobalTable"
)

type connector struct {
	kube client.Client
	opts []option
}

func (c *connector) Connect(ctx context.Context, cr *svcapitypes.GlobalTable) (managed.TypedExternalClient[*svcapitypes.GlobalTable], error) {
	sess, err := connectaws.GetConfigV1(ctx, c.kube, cr, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return newExternal(c.kube, svcapi.New(sess), c.opts), nil
}

func (e *external) Observe(ctx context.Context, cr *svcapitypes.GlobalTable) (managed.ExternalObservation, error) {
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	input := GenerateDescribeGlobalTableInput(cr)
	if err := e.preObserve(ctx, cr, input); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "pre-observe failed")
	}
	resp, err := e.client.DescribeGlobalTableWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err := e.lateInitialize(&cr.Spec.ForProvider, resp); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "late-init failed")
	}
	GenerateGlobalTable(resp).Status.AtProvider.DeepCopyInto(&cr.Status.AtProvider)
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

func (e *external) Create(ctx context.Context, cr *svcapitypes.GlobalTable) (managed.ExternalCreation, error) {
	cr.Status.SetConditions(xpv1.Creating())
	input := GenerateCreateGlobalTableInput(cr)
	if err := e.preCreate(ctx, cr, input); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "pre-create failed")
	}
	resp, err := e.client.CreateGlobalTableWithContext(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	if resp.GlobalTableDescription.CreationDateTime != nil {
		cr.Status.AtProvider.CreationDateTime = &metav1.Time{*resp.GlobalTableDescription.CreationDateTime}
	} else {
		cr.Status.AtProvider.CreationDateTime = nil
	}
	if resp.GlobalTableDescription.GlobalTableArn != nil {
		cr.Status.AtProvider.GlobalTableARN = resp.GlobalTableDescription.GlobalTableArn
	} else {
		cr.Status.AtProvider.GlobalTableARN = nil
	}
	if resp.GlobalTableDescription.GlobalTableName != nil {
		cr.Status.AtProvider.GlobalTableName = resp.GlobalTableDescription.GlobalTableName
	} else {
		cr.Status.AtProvider.GlobalTableName = nil
	}
	if resp.GlobalTableDescription.GlobalTableStatus != nil {
		cr.Status.AtProvider.GlobalTableStatus = resp.GlobalTableDescription.GlobalTableStatus
	} else {
		cr.Status.AtProvider.GlobalTableStatus = nil
	}
	if resp.GlobalTableDescription.ReplicationGroup != nil {
		f4 := []*svcapitypes.Replica{}
		for _, f4iter := range resp.GlobalTableDescription.ReplicationGroup {
			f4elem := &svcapitypes.Replica{}
			if f4iter.RegionName != nil {
				f4elem.RegionName = f4iter.RegionName
			}
			f4 = append(f4, f4elem)
		}
		cr.Spec.ForProvider.ReplicationGroup = f4
	} else {
		cr.Spec.ForProvider.ReplicationGroup = nil
	}

	return e.postCreate(ctx, cr, resp, managed.ExternalCreation{}, err)
}

func (e *external) Update(ctx context.Context, cr *svcapitypes.GlobalTable) (managed.ExternalUpdate, error) {
	input := GenerateUpdateGlobalTableInput(cr)
	if err := e.preUpdate(ctx, cr, input); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "pre-update failed")
	}
	resp, err := e.client.UpdateGlobalTableWithContext(ctx, input)
	return e.postUpdate(ctx, cr, resp, managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate))
}

func (e *external) Delete(ctx context.Context, cr *svcapitypes.GlobalTable) (managed.ExternalDelete, error) {
	cr.Status.SetConditions(xpv1.Deleting())
	return e.delete(ctx, cr)

}

func (e *external) Disconnect(ctx context.Context) error {
	// Unimplemented, required by newer versions of crossplane-runtime
	return nil
}

type option func(*external)

func newExternal(kube client.Client, client svcsdkapi.DynamoDBAPI, opts []option) *external {
	e := &external{
		kube:           kube,
		client:         client,
		preObserve:     nopPreObserve,
		postObserve:    nopPostObserve,
		lateInitialize: nopLateInitialize,
		isUpToDate:     alwaysUpToDate,
		preCreate:      nopPreCreate,
		postCreate:     nopPostCreate,
		delete:         nopDelete,
		preUpdate:      nopPreUpdate,
		postUpdate:     nopPostUpdate,
	}
	for _, f := range opts {
		f(e)
	}
	return e
}

type external struct {
	kube           client.Client
	client         svcsdkapi.DynamoDBAPI
	preObserve     func(context.Context, *svcapitypes.GlobalTable, *svcsdk.DescribeGlobalTableInput) error
	postObserve    func(context.Context, *svcapitypes.GlobalTable, *svcsdk.DescribeGlobalTableOutput, managed.ExternalObservation, error) (managed.ExternalObservation, error)
	lateInitialize func(*svcapitypes.GlobalTableParameters, *svcsdk.DescribeGlobalTableOutput) error
	isUpToDate     func(context.Context, *svcapitypes.GlobalTable, *svcsdk.DescribeGlobalTableOutput) (bool, string, error)
	preCreate      func(context.Context, *svcapitypes.GlobalTable, *svcsdk.CreateGlobalTableInput) error
	postCreate     func(context.Context, *svcapitypes.GlobalTable, *svcsdk.CreateGlobalTableOutput, managed.ExternalCreation, error) (managed.ExternalCreation, error)
	delete         func(context.Context, *svcapitypes.GlobalTable) (managed.ExternalDelete, error)
	preUpdate      func(context.Context, *svcapitypes.GlobalTable, *svcsdk.UpdateGlobalTableInput) error
	postUpdate     func(context.Context, *svcapitypes.GlobalTable, *svcsdk.UpdateGlobalTableOutput, managed.ExternalUpdate, error) (managed.ExternalUpdate, error)
}

func nopPreObserve(context.Context, *svcapitypes.GlobalTable, *svcsdk.DescribeGlobalTableInput) error {
	return nil
}

func nopPostObserve(_ context.Context, _ *svcapitypes.GlobalTable, _ *svcsdk.DescribeGlobalTableOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}
func nopLateInitialize(*svcapitypes.GlobalTableParameters, *svcsdk.DescribeGlobalTableOutput) error {
	return nil
}
func alwaysUpToDate(context.Context, *svcapitypes.GlobalTable, *svcsdk.DescribeGlobalTableOutput) (bool, string, error) {
	return true, "", nil
}

func nopPreCreate(context.Context, *svcapitypes.GlobalTable, *svcsdk.CreateGlobalTableInput) error {
	return nil
}
func nopPostCreate(_ context.Context, _ *svcapitypes.GlobalTable, _ *svcsdk.CreateGlobalTableOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}
func nopDelete(context.Context, *svcapitypes.GlobalTable) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, nil
}
func nopPreUpdate(context.Context, *svcapitypes.GlobalTable, *svcsdk.UpdateGlobalTableInput) error {
	return nil
}
func nopPostUpdate(_ context.Context, _ *svcapitypes.GlobalTable, _ *svcsdk.UpdateGlobalTableOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
