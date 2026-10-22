package templates

import rgdv1 "kro.run/resourcegraphdefinition/v1alpha1"

#CEL: string & =~"^\\$\\{.+\\}$"

#RGD: rgdv1.#ResourceGraphDefinition & {
	#config: #Config
	#name:   string
	#kind:   string
	metadata: {
		name: #name
		labels: #config.graphLabels & {
			"app.kubernetes.io/part-of": "krombat"
		}
	}
	spec: schema: {
		apiVersion: "v1alpha1"
		group:      "game.k8s.example"
		kind:       #kind
		scope:      "Namespaced"
	}
}

#ChildMetadata: {
	#name:     *"${schema.metadata.name}" | #CEL
	#dungeon:  *"${schema.spec.dungeonName}" | #CEL
	name:      #name
	namespace: "${schema.metadata.namespace}"
	labels: {
		"game.k8s.example/dungeon": #dungeon
		[string]:                   string
	}
}

#Resource: {
	id!: string
	forEach?: [...{[string]: #CEL}]
	includeWhen?: [...#CEL]
	readyWhen?: [...#CEL]
	template!: {...}
}

#StateResource: #Resource & {
	#entity: string
	#name:   *"${schema.metadata.name}" | #CEL
	let resourceName = #name
	template: {
		apiVersion: "v1"
		kind:       "ConfigMap"
		metadata: #ChildMetadata & {
			#name: resourceName
			labels: "game.k8s.example/entity": #entity
		}
		data: {[string]: string}
	}
}

#GameResource: #Resource & {
	#kind:    string
	#name:    #CEL
	#dungeon: *"${schema.spec.dungeonName}" | #CEL
	let resourceName = #name
	let parentDungeon = #dungeon
	template: {
		apiVersion: "game.k8s.example/v1alpha1"
		kind:       #kind
		metadata: #ChildMetadata & {
			#name:    resourceName
			#dungeon: parentDungeon
		}
		spec: {
			dungeonName: #dungeon
			...
		}
	}
}

#LootResource: #GameResource & {
	#kind: "Loot"
	#name: "${schema.metadata.name + '-loot'}"
	id:    "loot"
	includeWhen: ["${schema.spec.hp == 0}"]
}

#HPColumn: {
	name:     "HP"
	type:     "integer"
	jsonPath: ".spec.hp"
}

#StateColumn: {
	name:     "State"
	type:     "string"
	jsonPath: ".status.entityState"
}

#ReadyColumn: {
	name:     "Ready"
	type:     "string"
	jsonPath: ".status.conditions[?(@.type==\"Ready\")].status"
}
