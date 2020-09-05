// +build !ignore_autogenerated

/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImageRef) DeepCopyInto(out *ImageRef) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImageRef.
func (in *ImageRef) DeepCopy() *ImageRef {
	if in == nil {
		return nil
	}
	out := new(ImageRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresBackup) DeepCopyInto(out *PostgresBackup) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresBackup.
func (in *PostgresBackup) DeepCopy() *PostgresBackup {
	if in == nil {
		return nil
	}
	out := new(PostgresBackup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresRestore) DeepCopyInto(out *PostgresRestore) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresRestore.
func (in *PostgresRestore) DeepCopy() *PostgresRestore {
	if in == nil {
		return nil
	}
	out := new(PostgresRestore)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StashPostgres) DeepCopyInto(out *StashPostgres) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StashPostgres.
func (in *StashPostgres) DeepCopy() *StashPostgres {
	if in == nil {
		return nil
	}
	out := new(StashPostgres)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StashPostgres) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StashPostgresList) DeepCopyInto(out *StashPostgresList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StashPostgres, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StashPostgresList.
func (in *StashPostgresList) DeepCopy() *StashPostgresList {
	if in == nil {
		return nil
	}
	out := new(StashPostgresList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StashPostgresList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StashPostgresSpec) DeepCopyInto(out *StashPostgresSpec) {
	*out = *in
	out.Image = in.Image
	out.Backup = in.Backup
	out.Restore = in.Restore
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StashPostgresSpec.
func (in *StashPostgresSpec) DeepCopy() *StashPostgresSpec {
	if in == nil {
		return nil
	}
	out := new(StashPostgresSpec)
	in.DeepCopyInto(out)
	return out
}
