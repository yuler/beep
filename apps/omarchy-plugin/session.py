#!/usr/bin/env python3
"""Print the signed-in Beep session as JSON, without tokens.

Account slug matches what `beep list` uses: the configured account, or the
personal account when the CLI config has no account slug.
"""

import json
import subprocess
import sys
import urllib.error
import urllib.request


def load_config():
    raw = subprocess.check_output(
        ["beep", "config", "show", "--json", "--show-token"],
        text=True,
    )
    return json.loads(raw)


def fetch_me(server_url, token):
    request = urllib.request.Request(
        server_url.rstrip("/") + "/api/v1/me",
        headers={
            "Authorization": "Bearer " + token,
            "Accept": "application/json",
            "User-Agent": "beep-omarchy-plugin",
        },
    )
    with urllib.request.urlopen(request, timeout=10) as response:
        return json.loads(response.read().decode())


def resolve_account_slug(config, me):
    configured = (config.get("AccountSlug") or "").strip()
    if configured:
        return configured
    if not me:
        return ""
    accounts = me.get("accounts") or []
    for account in accounts:
        if account.get("personal") and account.get("slug"):
            return account["slug"]
    last = (me.get("last_account_slug") or "").strip()
    if last:
        return last
    if len(accounts) == 1:
        return accounts[0].get("slug") or ""
    return ""


def main():
    try:
        config = load_config()
    except Exception as err:
        print(json.dumps({"error": str(err)}))
        return 1

    server_url = (config.get("ServerURL") or "").rstrip("/")
    token = config.get("AccessToken") or ""
    me = None
    if server_url and token and not (config.get("AccountSlug") or "").strip():
        try:
            me = fetch_me(server_url, token)
        except (urllib.error.URLError, TimeoutError, json.JSONDecodeError, OSError):
            me = None

    print(json.dumps({
        "UserName": config.get("UserName") or "",
        "UserEmail": config.get("UserEmail") or "",
        "AccountSlug": resolve_account_slug(config, me),
        "ServerURL": server_url,
        "LoggedIn": bool(token),
    }))
    return 0


if __name__ == "__main__":
    sys.exit(main())
