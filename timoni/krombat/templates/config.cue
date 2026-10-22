package templates

import timoniv1 "timoni.sh/core/v1alpha1"

#Config: {
	graphLabels: timoniv1.#Labels
	dungeonDefaults: {
		monsters:   *3 | int & >=1 & <=10
		difficulty: *"easy" | "easy" | "normal" | "hard"
	}
}

#Instance: {
	config: #Config
	graphs: {
		loot: #LootGraph & {#config: config}
		hero: #HeroGraph & {#config: config}
		monster: #MonsterGraph & {#config: config}
		boss: #BossGraph & {#config: config}
		treasure: #TreasureGraph & {#config: config}
		dungeon: #DungeonGraph & {#config: config}
	}
}
