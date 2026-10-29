package main

runtime: {
	apiVersion: "v1alpha1"
	name:       "demo-infra"
	clusters: local: {
		group:       "demo"
		kubeContext: "minikube"
	}
	values: [{
		query: "k8s:v1:ConfigMap:demo:infra-values"
		for: "DEMO_REPLICAS": "obj.data.replicas"
	}, {
		query: "k8s:v1:Secret:demo:infra-secrets"
		for: "DEMO_TOKEN": "obj.data.token"
	}]
}
