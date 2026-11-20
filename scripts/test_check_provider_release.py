import argparse
import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import check_provider_release as release


class TestGoModuleUpdate(unittest.TestCase):
    @patch.object(release, "run")
    @patch.object(release, "get_release_commit", return_value="0123456789abcdef")
    def test_updates_module_at_release_commit(self, get_release_commit, run):
        release.update_go_module("v5.10.0")

        get_release_commit.assert_called_once_with("v5.10.0")
        run.assert_called_once_with(
            ["go", "get", f"{release.GO_MODULE}@0123456789abcdef"]
        )

    @patch.object(release, "run_gh", return_value="0123456789abcdef\n")
    def test_resolves_release_tag_to_commit(self, run_gh):
        commit = release.get_release_commit("v5.10.0")

        self.assertEqual(commit, "0123456789abcdef")
        run_gh.assert_called_once_with(
            [
                "api",
                f"repos/{release.UPSTREAM_REPO}/commits/v5.10.0",
                "--jq",
                ".sha",
            ]
        )

    def test_bump_success_code_is_distinct_from_unhandled_error(self):
        self.assertEqual(release.BUMPED_EXIT_CODE, 10)
        self.assertNotEqual(release.BUMPED_EXIT_CODE, 1)


PLATFORMS = ["linux_amd64", "linux_arm64"]
ALL_ASSETS = [
    "terraform-provider-keycloak_5.11.0_linux_amd64.zip",
    "terraform-provider-keycloak_5.11.0_linux_arm64.zip",
    "terraform-provider-keycloak_5.11.0_SHA256SUMS",
]


class TestReleaseAssets(unittest.TestCase):
    def test_reads_platforms_from_makefile(self):
        with tempfile.TemporaryDirectory() as d:
            makefile = Path(d) / "Makefile"
            makefile.write_text("PLATFORMS ?= linux_amd64 darwin_arm64\n")
            self.assertEqual(release.get_platforms(makefile), ["linux_amd64", "darwin_arm64"])

    def test_platforms_default_when_unset(self):
        with tempfile.TemporaryDirectory() as d:
            makefile = Path(d) / "Makefile"
            makefile.write_text("FOO ?= bar\n")
            self.assertEqual(release.get_platforms(makefile), release.DEFAULT_PLATFORMS)

    def test_no_missing_assets(self):
        self.assertEqual(release.missing_assets("5.11.0", ALL_ASSETS, PLATFORMS), [])

    def test_reports_missing_assets(self):
        self.assertEqual(
            release.missing_assets("5.11.0", ALL_ASSETS[:1], PLATFORMS),
            ["terraform-provider-keycloak_5.11.0_linux_arm64.zip"],
        )
        self.assertEqual(
            release.missing_assets("5.11.0", [], PLATFORMS),
            release.required_assets("5.11.0", PLATFORMS),
        )

    @patch.object(release, "run_gh")
    def test_latest_release_includes_assets(self, run_gh):
        run_gh.return_value = json.dumps(
            {"tag_name": "v5.11.0", "html_url": "u", "body": None, "assets": ALL_ASSETS}
        )
        self.assertEqual(release.get_latest_release(), ("v5.11.0", "u", "", ALL_ASSETS))


@patch.object(release, "get_platforms", return_value=PLATFORMS)
@patch.object(release, "get_current_version", return_value="5.10.0")
class TestSkipReleaseWithoutAssets(unittest.TestCase):
    args = argparse.Namespace(repo="o/r", dry_run=True, label="dependencies", pr_body_file=None)

    @patch.object(release, "fetch_open_issues")
    @patch.object(release, "get_latest_release", return_value=("v5.11.0", "u", "", []))
    def test_issue_skipped_without_assets(self, _latest, fetch_open_issues, *_):
        self.assertEqual(release.cmd_issue(self.args), 0)
        fetch_open_issues.assert_not_called()

    @patch.object(release, "fetch_open_issues", return_value=[])
    @patch.object(release, "get_latest_release", return_value=("v5.11.0", "u", "", ALL_ASSETS))
    def test_issue_proceeds_with_assets(self, _latest, fetch_open_issues, *_):
        self.assertEqual(release.cmd_issue(self.args), 0)
        fetch_open_issues.assert_called_once_with("o/r")

    @patch.object(release, "get_latest_release", return_value=("v5.11.0", "u", "", ALL_ASSETS[:1]))
    def test_bump_skipped_without_assets(self, *_):
        self.assertEqual(release.cmd_bump(self.args), 0)

    @patch.object(release, "get_latest_release", return_value=("v5.11.0", "u", "", ALL_ASSETS))
    def test_bump_proceeds_with_assets(self, *_):
        self.assertEqual(release.cmd_bump(self.args), release.BUMPED_EXIT_CODE)


if __name__ == "__main__":
    unittest.main()
