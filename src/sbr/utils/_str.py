from collections.abc import Generator


def strip_comments(text: str) -> Generator[str, None, None]:
    for line in text.splitlines():
        s: str
        s, _, _ = line.partition("#")
        s = s.strip()
        if s:
            yield s


def split_strip(text: str, sep: str | None = ",") -> list[str]:
    return [s.strip() for s in text.split(sep)]
