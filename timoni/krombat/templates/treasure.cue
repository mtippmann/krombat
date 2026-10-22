package templates

#TreasureGraph: #RGD & {
	#name: "treasure-graph"
	#kind: "Treasure"
	spec: {
		schema: {
			spec: {
				dungeonName: "string | required=true"
				opened:      "integer | default=0 minimum=0 maximum=1"
			}
			status: treasureState: "${treasureState.data.state}"
		}
		resources: [
			#StateResource & {
				id:      "treasureState"
				#entity: "treasure"
				#name:   "${schema.metadata.name + '-state'}"
				template: data: state: "${schema.spec.opened == 1 ? 'opened' : 'unopened'}"
			},
			#Resource & {
				id: "treasureSecret"
				includeWhen: ["${schema.spec.opened == 1}"]
				template: {
					apiVersion: "v1"
					kind:       "Secret"
					metadata: #ChildMetadata & {
						labels: "game.k8s.example/entity": "treasure"
					}
					type: "Opaque"
					stringData: reward: "${'Victory trophy for ' + schema.spec.dungeonName}"
				}
			},
		]
	}
}
