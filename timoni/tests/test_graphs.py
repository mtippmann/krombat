import copy
import difflib
import json
from pathlib import Path
import subprocess
import tempfile
import unittest

import yaml


ROOT = Path(__file__).resolve().parents[2]
MODULE = ROOT / "timoni/krombat"


def render(values=None):
    command = [
        "timoni", "build", "krombat-graphs", str(MODULE),
        "--namespace", "rpg-system", "--output", "json",
    ]
    if values is not None:
        command.extend(["--values", "-"])
    result = subprocess.run(
        command, input=json.dumps({"values": values}) if values is not None else None,
        text=True, capture_output=True, check=True,
    )
    objects = json.loads(result.stdout)["items"]
    names = [obj["metadata"]["name"] for obj in objects]
    if len(names) != len(set(names)):
        raise AssertionError(f"Duplicate graph names: {names}")
    return {obj["metadata"]["name"]: obj for obj in objects}


class GraphAuthoringTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.defaults = render()

    def assert_graphs_equal(self, expected, actual):
        before = json.dumps(expected, sort_keys=True, indent=2).splitlines()
        after = json.dumps(actual, sort_keys=True, indent=2).splitlines()
        self.assertEqual(expected, actual, "\n".join(difflib.unified_diff(
            before, after, fromfile="expected RGDs", tofile="rendered RGDs",
        )))

    def test_defaults_preserve_all_six_original_graphs(self):
        originals = {}
        for path in sorted((ROOT / "manifests/rgds").glob("*-graph.yaml")):
            graph = yaml.safe_load(path.read_text())
            # kro defaults this field; the imported CUE schema requires it explicitly.
            graph["spec"]["schema"].setdefault("scope", "Namespaced")
            originals[graph["metadata"]["name"]] = graph
        self.assertEqual(len(originals), 6)
        self.assert_graphs_equal(originals, self.defaults)

    def test_values_change_graph_contract_not_game_state(self):
        expected = copy.deepcopy(self.defaults)
        for graph in expected.values():
            graph["metadata"]["labels"]["platform.example/team"] = "workshop"
        spec = expected["dungeon-graph"]["spec"]["schema"]["spec"]
        spec["monsters"] = "integer | default=5 minimum=1 maximum=10"
        spec["difficulty"] = 'string | default="normal" enum=easy,normal,hard'
        actual = render({
            "graphLabels": {"platform.example/team": "workshop"},
            "dungeonDefaults": {"monsters": 5, "difficulty": "normal"},
        })
        self.assert_graphs_equal(expected, actual)

    def test_invalid_authoring_values_fail(self):
        for values, field in [
            ({"dungeonDefaults": {"monsters": 0}}, "monsters"),
            ({"dungeonDefaults": {"difficulty": "nightmare"}}, "difficulty"),
            ({"graphLabels": {"platform.example/team": "not a label"}}, "graphLabels"),
            ({"graphLabels": {"app.kubernetes.io/part-of": "other"}}, "part-of"),
        ]:
            with self.subTest(values=values):
                result = subprocess.run(
                    ["timoni", "build", "krombat-graphs", str(MODULE), "--values", "-"],
                    input=json.dumps({"values": values}), text=True, capture_output=True,
                )
                self.assertNotEqual(result.returncode, 0, result.stdout)
                self.assertIn(field, result.stderr)

    def test_module_packages_as_local_oci(self):
        with tempfile.TemporaryDirectory(prefix="krombat-timoni-") as directory:
            output = Path(directory) / "module"
            subprocess.run([
                "timoni", "mod", "build", str(MODULE), "--version", "0.1.0",
                "--format", "oci-layout", "--output", str(output),
            ], check=True, capture_output=True, text=True)
            index = json.loads((output / "index.json").read_text())
            algorithm, digest = index["manifests"][0]["digest"].split(":")
            manifest = json.loads((output / "blobs" / algorithm / digest).read_text())
            self.assertEqual(manifest["annotations"]["org.opencontainers.image.version"], "0.1.0")
            self.assertTrue(manifest["layers"])


if __name__ == "__main__":
    unittest.main()
