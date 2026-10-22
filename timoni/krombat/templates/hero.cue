package templates

#HeroGraph: #RGD & {
	#name: "hero-graph"
	#kind: "Hero"
	spec: {
		schema: {
			spec: {
				dungeonName: "string | required=true"
				hp:          "integer | default=200 minimum=0 maximum=200"
			}
			status: {
				entityState: "${heroState.data.entityState}"
				maxHP:       "${heroState.data.maxHP}"
				hp:          "${heroState.data.hp}"
			}
		}
		resources: [
			#StateResource & {
				id:      "heroState"
				#entity: "hero"
				template: data: {
					entityState: "${schema.spec.hp > 0 ? 'alive' : 'defeated'}"
					maxHP:       "200"
					hp:          "${string(schema.spec.hp)}"
				}
			},
		]
	}
}
