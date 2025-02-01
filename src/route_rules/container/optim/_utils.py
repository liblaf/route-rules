import itertools
import re


def split_domain(domain: str) -> list[str]:
    return re.split(r"(\.)", domain)


def domain_suffixes(domain: str) -> list[str]:
    labels: list[str] = split_domain(domain[::-1])
    suffixes: list[str] = [suffix[::-1] for suffix in itertools.accumulate(labels)]
    return suffixes


def match_domain_suffix(domain: str, suffix: set[str]) -> bool:
    return any((s in suffix) for s in domain_suffixes(domain))


def match_domain_keyword(domain: str, keyword: set[str]) -> bool:
    return any((k in domain) for k in keyword)
