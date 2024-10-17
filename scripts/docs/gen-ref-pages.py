"""Generate the code reference pages and navigation."""

from pathlib import Path

from mkdocs_gen_files.editor import FilesEditor
from mkdocs_gen_files.nav import Nav

editor: FilesEditor = FilesEditor.current()
nav = Nav()
root: Path = Path(__file__).parent.parent.parent
docs_dir: Path = root / "docs"
src_dir: Path = root / "src"
pkg_dir: Path = src_dir / "sbr"
for path in sorted(pkg_dir.rglob("*.py")):
    module_path: Path = path.relative_to(src_dir).with_suffix("")
    doc_path: Path = path.relative_to(src_dir).with_suffix(".md")
    full_doc_path: Path = Path("api", doc_path)
    parts: tuple[str, ...] = tuple(module_path.parts)
    if parts[-1] == "__init__":
        parts = parts[:-1]
        doc_path = doc_path.with_name("README.md")
        full_doc_path = full_doc_path.with_name("README.md")
    elif parts[-1].startswith("_"):
        continue
    nav_parts: tuple[str, ...] = tuple(
        ".".join(parts[:i]) for i in range(1, len(parts) + 1)
    )
    nav[nav_parts] = doc_path.as_posix()
    with editor.open(full_doc_path.as_posix(), "w") as fd:
        identifier: str = ".".join(parts)
        fd.write(f"""\
::: {identifier}
""")
    editor.set_edit_path(
        full_doc_path.as_posix(), (".." / path.relative_to(root)).as_posix()
    )
with editor.open("api/SUMMARY.md", "w") as nav_file:
    nav_file.writelines(nav.build_literate_nav())
