SHELL := /bin/bash 

# ============================================================================
# Testing running system ( run this on cmd line while app running )

# expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

# put 10,000 requests through a service to test metrics gathering, logging, the whole onion
# hey -m GET -c 100 -n 10000 http://localhost:3000/v1/users/1/2

# To generate a private/public key PEM file 
# openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048 
# openssl rsa -pubout -in private.pem -out public.pem
# ./action-admin genkey

# Testing Auth
# curl -il http://localhost:3000/v1/testauth

# Assumes we have a token from running `make admin`
# curl -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/testauth

# Also able to generate a token, after seeding, by querying /v1/users/token
# curl -H "Authorization:  Basic <base64 encoded username:password>" http://localhost:3000/v1/users/token
#                                   Sm9obiBLcm9la2VyOmdvcGhlcg==

# ============================================================================
# Seeding the dgraph database with curl

# curl -H "Content-Type: application/json" http://172.22.0.2:8080/admin/schema -XPOST -d $'
# type User {
#  id: ID!
#  name: String! @search(by: [exact])
# }'

# GraphQL Playground query example
# {
#   getUser(id:"0x9") {
#     name
#     email
#   }
# }

# ============================================================================

# --help shows the user usage options for cmd line flags 
# piping structured logging to logfmt tooling renders human-readable output
run:
	go run app/services/action-api/main.go
	# go run app/services/action-api/main.go | go run app/tooling/logfmt/main.go
	# go run app/services/action-api/main.go --help

admin: 
	go run app/services/action-admin/main.go gentoken "jnkroeker@gmail.com"
	# go run app/tooling/admin/main.go (deprecated: does not use a valid user id for subject in token) 

# ============================================================================
# Building containers 

VERSION := 0.1 

all: action-api 

action-api: 
	docker build \
		-f zarf/docker/dockerfile.action-api \
		-t action-api-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` .

# =========================================
# Modules support

tidy:
	go mod tidy
	go mod vendor

# =========================================
# Running from within k8s/kind

KIND_CLUSTER := action-cluster 

kind-up:
	kind create cluster \
		--image kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6 \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=action-system

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-load:
	# move into kind/action-pod folder, replace the image tag in kustomization.yaml with the version specified in this file 
	cd zarf/k8s/kind/action-pod; kustomize edit set images action-api-image=action-api-amd64:$(VERSION)
	kind load docker-image action-api-amd64:$(VERSION) --name $(KIND_CLUSTER)

kind-apply:
	# kustomize produces a yaml document that can be applied thru kubectl tooling
	# start from the kustomization.yaml in the kind/action-pod directory
	kustomize build zarf/k8s/kind/action-pod | kubectl apply -f -

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-logs:
	kubectl logs -l app=action --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go

kind-restart:
	kubectl rollout restart deployment action-pod

kind-update: all kind-load kind-restart

kind-update-apply: all kind-load kind-apply

kind-describe:
	kubectl describe nodes
	kubectl describe svc
	kubectl describe pod -l app=action

# Administration

schema:
	go run app/services/action-admin/main.go schema

seed: schema 
	go run app/services/action-admin/main.go seed

# Running tests within the local machine

test:
	# Find test files inside the entire project file structure
	# -count=1 ignores test cache and run all tests each time
	go test ./... -count=1
	staticcheck ./...