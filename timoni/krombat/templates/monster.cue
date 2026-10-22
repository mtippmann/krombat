package templates

#MonsterGraph: #RGD & {
	#name: "monster-graph"
	#kind: "Monster"
	spec: {
		schema: {
			additionalPrinterColumns: [#HPColumn, #StateColumn, #ReadyColumn]
			spec: {
				dungeonName: "string | required=true"
				index:       "integer | required=true minimum=0"
				hp:          "integer | required=true minimum=0"
			}
			status: {
				entityState: "${monsterState.data.entityState}"
				hp:          "${monsterState.data.hp}"
			}
		}
		resources: [
			#StateResource & {
				id:      "monsterState"
				#entity: "monster"
				template: data: {
					entityState: "${schema.spec.hp > 0 ? 'alive' : 'dead'}"
					hp:          "${string(schema.spec.hp)}"
				}
			},
			#LootResource & {
				template: spec: rarity: "${schema.spec.index % 2 == 0 ? 'rare' : 'common'}"
			},
		]
	}
}
