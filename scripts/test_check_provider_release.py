import unittest
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


if __name__ == "__main__":
    unittest.main()
