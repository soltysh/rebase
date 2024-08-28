package cmd

// 1. Add OpenShift dependencies
// go mod edit -require github.com/openshift/api@master \
//		-require github.com/openshift/client-go@master \
//		-require github.com/openshift/library-go@master \
//		-require github.com/openshift/apiserver-library-go@master \
//		-replace github.com/onsi/ginkgo/v2=github.com/soltysh/ginkgo/v2@v2.15-openshift-4.17

// 2. Run go mod tidy to resolve deps from 1

// 3. Run hack/update-vendor.sh.

// 4. make update OS_RUN_WITHOUT_DOCKER=yes

// for 1.30 changes:
// - hack/lib/golang.sh:
// https://github.com/openshift/kubernetes/commit/5deaeca317553053da30caac50c281562cf6375b#diff-a82169c65a556acd9e4ed51006b5cc12ad72a13e486c5e2daf6a4e869263450fR549

// - hack/update-vendor.sh:
// https://github.com/openshift/kubernetes/commit/5deaeca317553053da30caac50c281562cf6375b#diff-b6ed2d0e481e37c6d38a9c0da57141de5245f4a4b58ee5d49ca95f7f2f7b010dR360
