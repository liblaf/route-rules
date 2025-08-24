import enum


class Behavior(enum.StrEnum):
    DOMAIN = enum.auto()
    IPCIDR = enum.auto()
    CLASSICAL = enum.auto()


class Format(enum.StrEnum):
    YAML = enum.auto()
    TEXT = enum.auto()
    MRS = enum.auto()
