import hishel
import httpx
from mkdocs_gen_files import FilesEditor

import route_rules as rr

editor: FilesEditor = FilesEditor.current()


def main() -> None:
    dist_url: str = "https://raw.githubusercontent.com/liblaf/route-rules/mihomo"
    with hishel.CacheClient(follow_redirects=True) as client:
        response: httpx.Response = client.get(f"{dist_url}/meta.json")
        response = response.raise_for_status()
        meta: rr.Meta = rr.Meta.json_decode(response.text)
        for recipe in meta.recipes:
            with editor.open(f"rulesets/{recipe.slug}.md", "w") as fp:
                url: str = f"{dist_url}/docs/{recipe.slug}.md"
                response: httpx.Response = client.get(url)
                response = response.raise_for_status()
                fp.write(response.text)


main()
