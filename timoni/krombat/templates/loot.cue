package templates

#LootGraph: #RGD & {
	#name: "loot-graph"
	#kind: "Loot"
	spec: {
		schema: {
			spec: {
				dungeonName: "string | required=true"
				rarity:      "string | default=\"common\" enum=common,rare,epic"
			}
			status: {
				description: "${lootInfo.data.description}"
				itemName:    "${lootSecret.metadata.name}"
			}
		}
		resources: [
			#StateResource & {
				id:      "lootInfo"
				#entity: "loot"
				#name:   "${schema.metadata.name + '-info'}"
				template: data: description: "${'A ' + schema.spec.rarity + ' trophy (display only)'}"
			},
			#Resource & {
				id: "lootSecret"
				template: {
					apiVersion: "v1"
					kind:       "Secret"
					metadata: #ChildMetadata & {
						labels: "game.k8s.example/entity": "loot"
					}
					type: "Opaque"
					stringData: {
						rarity: "${schema.spec.rarity}"
						trophy: "${random.seededString(8, schema.metadata.name)}"
					}
				}
			},
		]
	}
}
