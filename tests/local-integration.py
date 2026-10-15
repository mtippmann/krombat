"""API and real-resource assertions. Gameplay browser journeys are separate."""

import http.cookiejar
import json
import os
import subprocess
import time
import urllib.error
import urllib.request
import uuid

assert os.environ.get("KUBECTL_CONTEXT", "minikube") == "minikube"
BASE = os.environ.get("BASE_URL", "http://krombat.test")
assert BASE == "http://krombat.test", "Local tests cannot target production"
NAME = "check-" + uuid.uuid4().hex[:10]
PATH = "/api/v1/dungeons/rpg-demo/" + NAME
client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
created = False


def api(method, path, body=None, expected=200, authenticated=True):
    data = json.dumps(body).encode() if body is not None else None
    request = urllib.request.Request(BASE + path, data=data, method=method,
                                     headers={"Content-Type": "application/json"})
    try:
        response = (client.open if authenticated else urllib.request.urlopen)(request, timeout=15)
    except urllib.error.HTTPError as error:
        response = error
    raw = response.read()
    assert response.code == expected, (method, path, response.code, raw.decode())
    return json.loads(raw) if raw else None


def kctl(*args):
    return subprocess.check_output(
        ["kubectl", "--context", "minikube", "-n", "rpg-demo", *args], text=True
    )


def resource(kind, name):
    return json.loads(kctl("get", kind, name, "-o", "json"))


def eventually(description, predicate):
    deadline = time.monotonic() + 90
    while time.monotonic() < deadline:
        if predicate():
            print("PASS", description)
            return
        time.sleep(1)
    raise AssertionError("Timed out: " + description)


def listed(kinds):
    return json.loads(kctl("get", kinds, "-l", "game.k8s.example/dungeon=" + NAME,
                          "-o", "json"))["items"]


def ready():
    dungeon = resource("dungeon", NAME)
    return any(c["type"] == "Ready" and c["status"] == "True"
               for c in dungeon.get("status", {}).get("conditions", []))


def attack(target):
    current = api("GET", PATH)
    time.sleep(0.35)
    return api("POST", PATH + "/attacks",
               {"target": target, "seq": current["spec"]["attackSeq"]})


try:
    api("GET", "/api/v1/auth/me", expected=401, authenticated=False)
    api("POST", "/api/v1/dungeons", {"name": NAME, "monsters": 2}, expected=401,
        authenticated=False)
    api("POST", "/api/v1/auth/login", {})
    assert api("GET", "/api/v1/auth/me")["login"] == "demo"
    api("POST", "/api/v1/dungeons",
        {"name": NAME, "monsters": 0, "difficulty": "easy", "heroClass": "warrior"}, expected=400)
    dungeon = api("POST", "/api/v1/dungeons",
                  {"name": NAME, "monsters": 2, "difficulty": "easy", "heroClass": "warrior"},
                  expected=201)
    created = True
    kctl("label", "dungeon", NAME, "krombat.io/test-run=" + NAME)
    assert dungeon["spec"]["game"]["heroHP"] == 200
    assert dungeon["spec"]["game"]["monsterHP"] == [30, 30]
    eventually("Dungeon and nested graphs become Ready", ready)
    children = listed("heroes,monsters,bosses,treasures")
    assert len([r for r in children if r["kind"] == "Monster"]) == 2
    parent_uid = dungeon["metadata"]["uid"]
    assert all(r["metadata"]["labels"].get("kro.run/instance-id") == parent_uid
               for r in children), "ApplySet child ownership must point to Dungeon UID"
    assert "kro.run/finalizer" in resource("dungeon", NAME)["metadata"]["finalizers"]
    assert not listed("loots,secrets"), "No loot or treasure before their conditions"
    print("PASS exact fan-out, ApplySet ownership/finalizer, conditional absence")

    api("POST", PATH + "/attacks", {"target": NAME + "-boss", "seq": 0}, expected=400)
    time.sleep(0.35)
    first = api("POST", PATH + "/attacks",
                {"target": NAME + "-monster-0", "seq": 0, "damage": 99999})
    assert first["spec"]["game"]["monsterHP"][0] == 5, "Client damage must be ignored"
    time.sleep(0.35)
    api("POST", PATH + "/attacks", {"target": NAME + "-monster-0", "seq": 0}, expected=409)
    print("PASS locked boss, server-owned damage and stale turn rejection")

    attack(NAME + "-monster-0")
    eventually("Monster kill creates nested loot Secret",
               lambda: any(r["metadata"]["name"] == NAME + "-monster-0-loot"
                           for r in listed("secrets")))
    kctl("patch", "dungeon", NAME, "--type=json",
         "-p", json.dumps([{"op": "replace", "path": "/spec/game/monsterHP/0", "value": 30}]))
    eventually("includeWhen prunes loot when its condition becomes false",
               lambda: not listed("loots,secrets"))
    attack(NAME + "-monster-0")
    attack(NAME + "-monster-0")
    attack(NAME + "-monster-1")
    attack(NAME + "-monster-1")
    eventually("Boss unlocks via kro aggregation",
               lambda: resource("dungeon", NAME).get("status", {}).get("bossState") == "ready")

    hero_cm = NAME + "-hero"
    kctl("patch", "configmap", hero_cm, "--type=merge",
         "-p", json.dumps({"data": {"hp": "9999"}}))
    eventually("kro restores edited child ConfigMap",
               lambda: resource("configmap", hero_cm)["data"]["hp"] ==
               str(resource("dungeon", NAME)["spec"]["game"]["heroHP"]))
    previous_uid = resource("configmap", hero_cm)["metadata"]["uid"]
    kctl("delete", "configmap", hero_cm, "--wait=true")
    eventually("kro recreates deleted child ConfigMap",
               lambda: any(r["metadata"]["name"] == hero_cm and
                           r["metadata"]["uid"] != previous_uid for r in listed("configmaps")))

    for hp, phase in [(100, "phase2"), (50, "phase3"), (0, "defeated")]:
        while api("GET", PATH)["spec"]["game"]["bossHP"] > hp:
            attack(NAME + "-boss")
        eventually("Boss threshold " + str(hp) + " -> " + phase,
                   lambda: resource("dungeon", NAME).get("status", {}).get("bossPhase") == phase)
    assert api("GET", PATH)["spec"]["game"]["heroHP"] > 0
    attack("open-treasure")
    eventually("Opening treasure creates conditional Secret",
               lambda: any(r["metadata"]["name"] == NAME + "-treasure"
                           for r in listed("secrets")))
    inspected = api("GET", PATH + "/resources")
    assert any(r["kind"] == "Monster" for r in inspected)
    assert any(r["kind"] == "Secret" for r in inspected)
    assert all(r["metadata"]["namespace"] == "rpg-demo" for r in inspected)
    result = api("POST", PATH + "/cel-eval", {"expr": "schema.spec.game.bossHP == 0"})
    assert result["result"] == "true", result
    print("PASS real resource inspection and upstream CEL playground")

    api("DELETE", PATH, expected=204)
    created = False
    eventually("Dungeon deletion cleans owned children",
               lambda: not listed("heroes,monsters,bosses,treasures,loots,configmaps,secrets"))
    print("All local API/graph acceptance assertions passed.")
finally:
    if created:
        assert resource("dungeon", NAME)["metadata"]["labels"]["krombat.io/test-run"] == NAME
        kctl("delete", "dungeon", NAME, "--ignore-not-found", "--wait=true", "--timeout=90s")
