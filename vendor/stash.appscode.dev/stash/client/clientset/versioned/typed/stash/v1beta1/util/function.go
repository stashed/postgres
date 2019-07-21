package util

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
	api "stash.appscode.dev/stash/apis/stash/v1beta1"
	cs "stash.appscode.dev/stash/client/clientset/versioned/typed/stash/v1beta1"
)

func CreateOrPatchFunction(c cs.StashV1beta1Interface, meta metav1.ObjectMeta, transform func(fn *api.Function) *api.Function) (*api.Function, kutil.VerbType, error) {
	cur, err := c.Functions().Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Function %s/%s.", meta.Namespace, meta.Name)
		out, err := c.Functions().Create(transform(&api.Function{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Function",
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchFunction(c, cur, transform)
}

func PatchFunction(c cs.StashV1beta1Interface, cur *api.Function, transform func(*api.Function) *api.Function) (*api.Function, kutil.VerbType, error) {
	return PatchFunctionObject(c, cur, transform(cur.DeepCopy()))
}

func PatchFunctionObject(c cs.StashV1beta1Interface, cur, mod *api.Function) (*api.Function, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonpatch.CreateMergePatch(curJson, modJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching Function %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.Functions().Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateFunction(c cs.StashV1beta1Interface, meta metav1.ObjectMeta, transform func(*api.Function) *api.Function) (result *api.Function, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.Functions().Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.Functions().Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Function %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update Function %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}
