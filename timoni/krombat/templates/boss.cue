package templates

#BossGraph: #RGD & {
	#name: "boss-graph"
	#kind: "Boss"
	spec: {
		schema: {
			additionalPrinterColumns: [
				#HPColumn,
				#StateColumn,
				{name: "Phase", type: "string", jsonPath: ".status.phase"},
				#ReadyColumn,
			]
			spec: {
				dungeonName:   "string | required=true"
				hp:            "integer | required=true minimum=0"
				maxHP:         "integer | default=200 minimum=1"
				monstersAlive: "integer | default=1 minimum=0"
			}
			status: {
				entityState:      "${bossState.data.entityState}"
				phase:            "${bossState.data.phase}"
				damageMultiplier: "${bossState.data.damageMultiplier}"
			}
		}
		resources: [
			#StateResource & {
				id:      "bossState"
				#entity: "boss"
				template: data: {
					hp:               "${string(schema.spec.hp)}"
					entityState:      "${schema.spec.hp == 0 ? 'defeated' : schema.spec.monstersAlive == 0 ? 'ready' : 'pending'}"
					phase:            "${schema.spec.hp == 0 ? 'defeated' : schema.spec.hp * 100 > schema.spec.maxHP * 50 ? 'phase1' : schema.spec.hp * 100 > schema.spec.maxHP * 25 ? 'phase2' : 'phase3'}"
					damageMultiplier: "${schema.spec.hp * 100 > schema.spec.maxHP * 50 ? '10' : schema.spec.hp * 100 > schema.spec.maxHP * 25 ? '13' : '16'}"
				}
			},
			#LootResource & {
				template: spec: rarity: "epic"
			},
		]
	}
}
