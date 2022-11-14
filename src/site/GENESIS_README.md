## Genesis of PrimaryIO site operator code base
Operator SDK (version 1.25.0, go: version go1.19.2) was used to generate site operator code base with 1 CRD : Site

Reference : https://sdk.operatorframework.io/docs/building-operators/golang/

#1. Generate operator code base
operator-sdk init --domain primaryio.com --repo github.com/CacheboxInc/DRaaS/src/site

#2. Create a new API and Controller for custom resource Site
operator-sdk create api --group site --version v1alpha1 --kind Site --resource --controller
