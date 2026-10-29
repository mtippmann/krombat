package main

import rgdv1 "kro.run/resourcegraphdefinition/v1alpha1"

// Timoni requires the tags even though the RGD is cluster-scoped.
_name:      string @tag(name)
_namespace: string @tag(namespace)

// Platform release input: the default for new WebApp instances.
#Config: {
	apiDefaultReplicas: *2 | (int & >=1 & <=10)
}
values: #Config

// The shared RGD envelope can be reused for other platform APIs.
#RGD: rgdv1.#ResourceGraphDefinition & {
	metadata: labels: "platform.example/managed-by": "platform"
	spec: schema: {
		apiVersion: "v1alpha1"
		group:      "platform.example"
		scope:      "Namespaced"
	}
}

#WebAppRGD: #RGD & {
	metadata: name: "web-app"
	spec: {
		schema: {
			kind: "WebApp"
			spec: replicas: "integer | default=\(values.apiDefaultReplicas) minimum=1 maximum=10"
		}
		resources: [{
			id: "webDeployment"
			template: {
				apiVersion: "apps/v1"
				kind:       "Deployment"
				metadata: {
					name:      "${schema.metadata.name}"
					namespace: "${schema.metadata.namespace}"
				}
				spec: {
					replicas: "${schema.spec.replicas}"
					selector: matchLabels: app: "${schema.metadata.name}"
					template: {
						metadata: labels: app: "${schema.metadata.name}"
						spec: containers: [{
							name:  "web"
							image: "nginx:1.27"
							ports: [{containerPort: 80}]
						}]
					}
				}
			}
		}]
	}
}

timoni: {
	apiVersion: "v1alpha1"
	instance: config: values
	apply: graphs: [#WebAppRGD]
}
