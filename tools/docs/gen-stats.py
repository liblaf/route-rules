import httpx
from mkdocs_gen_files.editor import FilesEditor

editor: FilesEditor = FilesEditor.current()
resp: httpx.Response = httpx.get(
    "https://github.com/liblaf/sing-box-rules/raw/sing/README.md", follow_redirects=True
)
resp.raise_for_status()
with editor.open("stats.md", "w") as fp:
    fp.write(resp.text)
