package templates

#DungeonGraph: #RGD & {
	#config: #Config
	#name:   "dungeon-graph"
	#kind:   "Dungeon"
	spec: {
		schema: {
			spec: {
				monsters:   "integer | default=\(#config.dungeonDefaults.monsters) minimum=1 maximum=10"
				difficulty: "string | default=\"\(#config.dungeonDefaults.difficulty)\" enum=easy,normal,hard"
				heroClass:  "string | default=\"warrior\" enum=warrior"
				attackSeq:  "integer | default=0 minimum=0"
				game: {
					heroHP:         "integer | default=200 minimum=0 maximum=200"
					bossHP:         "integer | default=200 minimum=0"
					monsterHP:      "[]integer"
					treasureOpened: "integer | default=0 minimum=0 maximum=1"
				}
			}
			additionalPrinterColumns: [
				{name: "Monsters", type: "integer", jsonPath: ".status.livingMonsters"},
				{name: "Boss", type: "string", jsonPath: ".status.bossState"},
				{name: "Phase", type: "string", jsonPath: ".status.bossPhase"},
				{name: "HeroHP", type: "integer", jsonPath: ".spec.game.heroHP"},
				#ReadyColumn,
			]
			status: {
				livingMonsters: "${size(monsters.filter(m, m.status.entityState == 'alive'))}"
				bossState:      "${boss.status.entityState}"
				bossPhase:      "${boss.status.phase}"
				victory:        "${schema.spec.game.bossHP == 0 && schema.spec.game.heroHP > 0}"
				maxHeroHP:      "${hero.status.maxHP}"
				maxMonsterHP:   "${config.data.maxMonsterHP}"
				maxBossHP:      "${config.data.maxBossHP}"
				treasureState:  "${treasure.status.treasureState}"
			}
		}
		resources: [
			#Resource & {
				id: "config"
				template: {
					apiVersion: "v1"
					kind:       "ConfigMap"
					metadata: #ChildMetadata & {
						#name:    "${schema.metadata.name + '-config'}"
						#dungeon: "${schema.metadata.name}"
					}
					data: {
						maxMonsterHP: "${schema.spec.difficulty == 'easy' ? '30' : schema.spec.difficulty == 'hard' ? '80' : '50'}"
						maxBossHP:    "${schema.spec.difficulty == 'easy' ? '200' : schema.spec.difficulty == 'hard' ? '800' : '400'}"
						engine:       "Go computes turns; vanilla kro projects resources and derives entity state."
					}
				}
			},
			#GameResource & {
				id:       "hero"
				#kind:    "Hero"
				#name:    "${schema.metadata.name + '-hero'}"
				#dungeon: "${schema.metadata.name}"
				readyWhen: ["${hero.status.maxHP == '200'}"]
				template: spec: hp: "${schema.spec.game.heroHP}"
			},
			#GameResource & {
				id:       "monsters"
				#kind:    "Monster"
				#name:    "${schema.metadata.name + '-monster-' + string(idx)}"
				#dungeon: "${schema.metadata.name}"
				forEach: [{idx: "${lists.range(size(schema.spec.game.monsterHP))}"}]
				template: spec: {
					index: "${idx}"
					hp:    "${schema.spec.game.monsterHP[idx]}"
				}
			},
			#GameResource & {
				id:       "boss"
				#kind:    "Boss"
				#name:    "${schema.metadata.name + '-boss'}"
				#dungeon: "${schema.metadata.name}"
				template: spec: {
					hp:            "${schema.spec.game.bossHP}"
					maxHP:         "${int(config.data.maxBossHP)}"
					monstersAlive: "${size(schema.spec.game.monsterHP.filter(hp, hp > 0))}"
				}
			},
			#GameResource & {
				id:       "treasure"
				#kind:    "Treasure"
				#name:    "${schema.metadata.name + '-treasure'}"
				#dungeon: "${schema.metadata.name}"
				template: spec: opened: "${schema.spec.game.treasureOpened}"
			},
		]
	}
}
