"""Generate the code reference pages and navigation."""

import itertools
from pathlib import Path

import git
from mkdocs_gen_files.editor import FilesEditor
from mkdocs_gen_files.nav import Nav


def git_root() -> Path:
    repo = git.Repo(search_parent_directories=True)
    return Path(repo.working_dir)


def is_private(parts: tuple[str, ...]) -> bool:
    return any(part.startswith("_") for part in parts)


def main() -> None:
    # [Recipes - mkdocstrings](https://mkdocstrings.github.io/recipes/)
    root: Path = git_root()
    editor: FilesEditor = FilesEditor.current()
    nav = Nav()
    src_dir: Path = root / "src"
    for path in sorted(src_dir.rglob("*.py")):
        relative_path: Path = path.relative_to(src_dir)
        doc_path: Path = relative_path.with_suffix(".md")
        parts: tuple[str, ...] = relative_path.with_suffix("").parts
        if parts[-1] == "__init__":
            parts = parts[:-1]
            doc_path = doc_path.with_name("README.md")
        if is_private(parts):
            continue
        nav_parts: tuple[str, ...] = tuple(
            itertools.accumulate(parts, lambda x, y: x + "." + y)
        )
        nav[nav_parts] = doc_path.as_posix()
        full_doc_path = Path("api", doc_path)
        with editor.open(full_doc_path.as_posix(), "w") as fd:
            identifier: str = nav_parts[-1]
            fd.write(f"""::: {identifier}""")
        editor.set_edit_path(
            full_doc_path.as_posix(), (".." / path.relative_to(root)).as_posix()
        )
    with editor.open("api/SUMMARY.md", "w") as nav_file:
        nav_file.writelines(nav.build_literate_nav())


main()
