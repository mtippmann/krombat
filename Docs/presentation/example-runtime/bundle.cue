package main

#InfraValues: {
	replicas: int & >=1 & <=10
	token:    string & !=""
}

// Offline teaching contract. The module URL is deliberately a placeholder.
bundle: {
	_replicas: number @timoni(runtime:number:DEMO_REPLICAS)
	_token:    string @timoni(runtime:string:DEMO_TOKEN)

	apiVersion: "v1alpha1"
	name:       "infra-contract"
	instances: web: {
		module: {
			url:     "oci://example.invalid/modules/web"
			version: "0.0.0"
		}
		namespace: "demo"
		values: #InfraValues & {
			replicas: _replicas
			token:    _token
		}
	}
}
