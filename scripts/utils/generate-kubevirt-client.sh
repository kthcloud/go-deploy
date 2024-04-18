go get kubevirt.io/api
go install k8s.io/code-generator/cmd/client-gen@latest

# create boilerplate.go.txt
echo "" > boilerplate.go.txt
client-gen --input-base="kubevirt.io/api/" --input="core/v1" --input="snapshot/v1alpha1" --output-package="go-deploy/pkg/imp/kubevirt" --output-base="../../" --clientset-name="kubevirt" --go-header-file boilerplate.go.txt
rm boilerplate.go.txt

