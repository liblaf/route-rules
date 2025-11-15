import enum


class Behavior(enum.StrEnum):
    DOMAIN = enum.auto()
    IPCIDR = enum.auto()
    CLASSICAL = enum.auto()


class Format(enum.StrEnum):
    YAML = enum.auto()
    TEXT = enum.auto()
    MRS = enum.auto()

    @property
    def ext(self) -> str:
        return _FORMAT_EXT[self]


_FORMAT_EXT: dict[Format, str] = {
    Format.YAML: ".yaml",
    Format.TEXT: ".list",
    Format.MRS: ".mrs",
}
