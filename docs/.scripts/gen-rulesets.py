import hishel
import httpx
from mkdocs_gen_files import FilesEditor

import route_rules as rr

editor: FilesEditor = FilesEditor.current()


def main() -> None:
    with hishel.CacheClient(follow_redirects=True) as client:
        response: httpx.Response = client.get(
            "https://raw.githubusercontent.com/liblaf/route-rules/dist/meta.json"
        ).raise_for_status()
        meta: rr.Meta = rr.Meta.json_decode(response.text)
    for recipe in meta.recipes:
        with editor.open(f"rulesets/{recipe.slug}.md", "w") as fp:
            url: str = f"https://raw.githubusercontent.com/liblaf/route-rules/dist/docs/{recipe.slug}.md"
            fp.write(f'{{ include-markdown "{url}" }}\n')


main()
