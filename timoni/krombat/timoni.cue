package main

import templates "timoni.sh/krombat/templates"

values: templates.#Config

// Timoni requires these tags even when a module only emits cluster-scoped RGDs.
_instanceName: string @tag(name)
_namespace:    string @tag(namespace)

timoni: {
	apiVersion: "v1alpha1"
	instance: templates.#Instance & {config: values}
	apply: graphs: [
		instance.graphs.loot,
		instance.graphs.hero,
		instance.graphs.monster,
		instance.graphs.boss,
		instance.graphs.treasure,
		instance.graphs.dungeon,
	]
}
