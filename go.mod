module stash.appscode.dev/postgres

go 1.12

require (
	github.com/appscode/go v0.0.0-20190621064509-6b292c9166e3
	github.com/gophercloud/gophercloud v0.0.0-20190520235722-e87e5f90e7e6 // indirect
	github.com/spf13/cobra v0.0.5
	k8s.io/api v0.0.0-20190515023547-db5a9d1c40eb // indirect
	k8s.io/apimachinery v0.0.0-20190515023456-b74e4c97951f
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.3.1 // indirect
	k8s.io/kube-openapi v0.0.0-20190510232812-a01b7d5d6c22 // indirect
	k8s.io/kubernetes v1.14.2 // indirect
	k8s.io/utils v0.0.0-20190520173318-324c5df7d3f0 // indirect
	kmodules.xyz/client-go v0.0.0-20190715080709-7162a6c90b04
	kmodules.xyz/custom-resources v0.0.0-20190508103408-464e8324c3ec
	kmodules.xyz/objectstore-api v0.0.0-20190718002052-da668b440b0b // indirect
	stash.appscode.dev/stash v0.0.0-20190718155146-3de534baa0a0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.3.0+incompatible
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190508045248-a52a97a7a2bf
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190508082252-8397d761d4b5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190314001948-2899ed30580f
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190314002645-c892ea32361a
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190314000054-4a91899592f4
	k8s.io/klog => k8s.io/klog v0.3.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190314000639-da8327669ac5
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/kubernetes => k8s.io/kubernetes v1.14.0
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190314001731-1bd6a4002213
	k8s.io/utils => k8s.io/utils v0.0.0-20190221042446-c2654d5206da
)
