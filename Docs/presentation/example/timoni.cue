package main

// Timoni supplies these fields from the instance name and namespace.
_name:      string @tag(name)
_namespace: string @tag(namespace)

#Config: {
	replicas: *2 | (int & >=1 & <=10)
}
values: #Config

#Deployment: {
	apiVersion: "apps/v1"
	kind:       "Deployment"
	metadata: {
		name:      _name
		namespace: _namespace
	}
	spec: {
		replicas: values.replicas
		selector: matchLabels: app: _name
		template: {
			metadata: labels: app: _name
			spec: containers: [{
				name:  "web"
				image: "nginx:1.27"
				ports: [{containerPort: 80}]
			}]
		}
	}
}

timoni: {
	apiVersion: "v1alpha1"
	instance: config: values
	apply: resources: [#Deployment]
}
