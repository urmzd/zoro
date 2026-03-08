"""Minimal SearXNG entry point for PyInstaller bundling."""

import os
import sys

# When running as a PyInstaller bundle, _MEIPASS points to the temp extract dir.
# Use it to locate the bundled settings.yml.
if getattr(sys, "_MEIPASS", None):
    base = sys._MEIPASS
else:
    base = os.path.dirname(os.path.abspath(__file__))

os.environ["SEARXNG_SETTINGS_PATH"] = os.path.join(base, "settings.yml")

from searx.webapp import app  # noqa: E402

if __name__ == "__main__":
    app.run(host="127.0.0.1", port=8888)
